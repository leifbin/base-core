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

func LoadNacosConfig[T any](cfg EnvConfig, dest *T, onConfigChange func(*T, []ConfigDiff)) error {
	client := getNacosClient()
	if client == nil {
		if err := InitNacosClient(cfg); err != nil {
			return err
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
		return err
	}

	err = yaml.Unmarshal([]byte(content), dest)
	if err != nil {
		slog.Error("解析配置失败", "err", err)
		return err
	}

	// 初始化 lastFullConfig
	setLastFullConfig(dest)

	// 定义防抖回调包装结构
	type debouncedData struct {
		Config *T
		Diffs  []ConfigDiff
	}

	debouncer := NewDebouncer(20*time.Second, func(data debouncedData) {
		setLastFullConfig(data.Config)
		slog.Debug(">>>防抖触发配置变更回调>>>", "diffs_count", len(data.Diffs))
		onConfigChange(data.Config, data.Diffs)
	})

	err = nacosClient.ListenConfig(vo.ConfigParam{
		DataId: cfg.DATA_ID,
		Group:  cfg.GROUP,
		OnChange: func(namespace, group, dataId, data string) {
			var newConfig T
			if unmarshalErr := yaml.Unmarshal([]byte(data), &newConfig); unmarshalErr != nil {
				UpdateNacosHealth(false, "Config parse error: "+unmarshalErr.Error())
				slog.Error("解析更新后的配置失败", "err", unmarshalErr)
				return // 不要去覆盖 lastFullConfig
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
		return err
	}
	StartNacosHealthMonitor(getNacosClient(), cfg)
	return nil
}
