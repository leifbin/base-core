package config

import (
	"sync"
	"time"

	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

var NacosHealth struct {
	Healthy   bool
	LastCheck time.Time
	Error     string
	Mutex     sync.RWMutex
}

func UpdateNacosHealth(healthy bool, errMsg string) {
	NacosHealth.Mutex.Lock()
	defer NacosHealth.Mutex.Unlock()

	NacosHealth.Healthy = healthy
	NacosHealth.LastCheck = time.Now()
	NacosHealth.Error = errMsg
}

func StartNacosHealthMonitor(configClient config_client.IConfigClient, cfg EnvConfig) {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			// 通过获取配置检查Nacos连接状态
			_, err := configClient.GetConfig(vo.ConfigParam{
				DataId: cfg.DATA_ID,
				Group:  cfg.GROUP,
			})

			if err != nil {
				UpdateNacosHealth(false, "Nacos connection failed: "+err.Error())
			} else {
				UpdateNacosHealth(true, "")
			}
		}
	}()
}
