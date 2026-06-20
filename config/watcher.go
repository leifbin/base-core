package config

import (
	"log/slog"
	"time"

	"github.com/nacos-group/nacos-sdk-go/vo"
	"gopkg.in/yaml.v2"
)

// NewWatcher 创建一个 Nacos 配置监听器。
// 每个 DataId 需要创建一个独立的 Watcher。
func NewWatcher[T any](cfg EnvConfig) *Watcher[T] {
	return &Watcher[T]{
		envCfg: cfg,
		stopCh: make(chan struct{}),
	}
}

// Load 加载初始配置并开始监听变更。
// 返回一个清理函数，用于停止监听和释放资源。
func (w *Watcher[T]) Load(dest *T, onConfigChange func(*T, []ConfigDiff)) (func(), error) {
	client := getNacosClient()
	if client == nil {
		if err := InitNacosClient(w.envCfg); err != nil {
			return nil, err
		}
		client = getNacosClient()
	}

	// 拉取初始配置
	content, err := client.GetConfig(vo.ConfigParam{
		DataId: w.envCfg.NACOS_DATA_ID,
		Group:  w.envCfg.NACOS_GROUP,
	})
	slog.Debug("Nacos 返回的原始内容", "length", len(content), "content", content)
	StartNacosHealthMonitor(client, w.envCfg)
	if err != nil {
		slog.Error("获取配置失败", "dataId", w.envCfg.NACOS_DATA_ID, "err", err)
		return nil, err
	}

	if err := yaml.Unmarshal([]byte(content), dest); err != nil {
		slog.Error("解析配置失败", "dataId", w.envCfg.NACOS_DATA_ID, "err", err)
		return nil, err
	}

	// 保存初始快照
	w.mu.Lock()
	w.lastConfig = dest
	w.mu.Unlock()

	// 创建防抖器
	w.debouncer = NewDebouncer(20*time.Second, func(data *watcherData[T]) {
		w.mu.Lock()
		w.lastConfig = data.config
		w.mu.Unlock()
		slog.Debug(">>>防抖触发配置变更回调>>>", "dataId", w.envCfg.NACOS_DATA_ID, "diffs_count", len(data.diffs))
		onConfigChange(data.config, data.diffs)
	})

	// 注册 Nacos 监听
	err = client.ListenConfig(vo.ConfigParam{
		DataId: w.envCfg.NACOS_DATA_ID,
		Group:  w.envCfg.NACOS_GROUP,
		OnChange: func(NACOS_NAMESPACE, NACOS_GROUP, dataId, data string) {
			var newConfig T
			if unmarshalErr := yaml.Unmarshal([]byte(data), &newConfig); unmarshalErr != nil {
				UpdateNacosHealth(false, "Config parse error: "+unmarshalErr.Error())
				slog.Error("解析更新后的配置失败", "dataId", dataId, "err", unmarshalErr)
				return
			}
			w.mu.RLock()
			old := w.lastConfig
			w.mu.RUnlock()

			changed, diffs := CompareConfigs(old, &newConfig)
			if changed {
				slog.Debug(">>>配置变更（进入防抖通道）>>>", "dataId", dataId)
				w.debouncer.Submit(&watcherData[T]{config: &newConfig, diffs: diffs})
			} else {
				slog.Warn("配置内容未发生实质性变化，跳过回调", "dataId", dataId)
			}
		},
	})
	if err != nil {
		slog.Error("监听配置失败", "dataId", w.envCfg.NACOS_DATA_ID, "err", err)
		return nil, err
	}

	// 返回清理函数
	w.cleanup = func() {
		close(w.stopCh)
		if w.debouncer != nil {
			w.debouncer.Stop()
		}
	}
	return w.cleanup, nil
}
