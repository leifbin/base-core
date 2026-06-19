package config

import (
	"log/slog"
	"sync"
	"time"

	"github.com/nacos-group/nacos-sdk-go/vo"
	"gopkg.in/yaml.v2"
)

// 存储上一次的完整配置
var (
	lastFullConfig interface{}
	cfgLock        sync.RWMutex
)

// GetLastFullConfig 返回上一次成功加载的完整配置的副本。
// 用于在配置变更时对比差异。
func GetLastFullConfig() interface{} {
	cfgLock.RLock()
	defer cfgLock.RUnlock()
	return lastFullConfig
}

func setLastFullConfig(cfg interface{}) {
	cfgLock.Lock()
	defer cfgLock.Unlock()
	lastFullConfig = cfg
}

// LoadNacosConfig 从 Nacos 加载 YAML 配置并反序列化为指定类型 T，
// 同时注册监听器实时跟踪配置变更。
// 变更时会计算差异并通过 onConfigChange 回调通知调用方，
// 回调带有 20 秒防抖机制，避免频繁触发。
func LoadNacosConfig[T any](cfg EnvConfig, dest *T, onConfigChange func(*T, []ConfigDiff)) (func(), error) {
	client := getNacosClient()
	if client == nil {
		if err := InitNacosClient(cfg); err != nil {
			return nil, err // ← 改
		}
		client = getNacosClient()
	}
	content, err := client.GetConfig(vo.ConfigParam{
		DataId: cfg.DATA_ID,
		Group:  cfg.GROUP,
	})

	UpdateNacosHealth(err == nil, "")
	if err != nil {
		slog.Error("获取配置失败", "err", err)
		return nil, err // ← 改
	}

	err = yaml.Unmarshal([]byte(content), dest)
	if err != nil {
		slog.Error("解析配置失败", "err", err)
		return nil, err // ← 改
	}

	setLastFullConfig(dest)

	type debouncedData struct {
		Config *T
		Diffs  []ConfigDiff
	}

	debouncer := NewDebouncer(20*time.Second, func(data debouncedData) {
		setLastFullConfig(data.Config)
		slog.Debug(">>>防抖触发配置变更回调>>>", "diffs_count", len(data.Diffs))
		onConfigChange(data.Config, data.Diffs)
	})

	err = getNacosClient().ListenConfig(vo.ConfigParam{ // ← 改：用 getNacosClient()
		DataId: cfg.DATA_ID,
		Group:  cfg.GROUP,
		OnChange: func(namespace, group, dataId, data string) {
			var newConfig T
			if unmarshalErr := yaml.Unmarshal([]byte(data), &newConfig); unmarshalErr != nil {
				UpdateNacosHealth(false, "Config parse error: "+unmarshalErr.Error())
				slog.Error("解析更新后的配置失败", "err", unmarshalErr)
				return
			}
			changed, diffs := CompareConfigs(GetLastFullConfig(), &newConfig)
			if changed {
				slog.Debug(">>>配置变更（进入防抖通道）>>>", slog.Bool("CompareConfigs", changed))
				debouncer.Submit(debouncedData{Config: &newConfig, Diffs: diffs})
			} else {
				slog.Warn("配置内容未发生实质性变化，跳过回调")
			}
		},
	})
	if err != nil {
		slog.Error("监听配置失败", "err", err)
		return nil, err // ← 改
	}

	stopCh := make(chan struct{})
	StartNacosHealthMonitor(getNacosClient(), cfg, stopCh)

	return func() {
		close(stopCh)
		debouncer.Stop()
	}, nil
}
