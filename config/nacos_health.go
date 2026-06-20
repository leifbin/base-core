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

var healthOnce sync.Once

// StartNacosHealthMonitor 启动后台 goroutine 定期检查 Nacos 连接状态。
// 全局只会启动一个实例，无论被调用多少次。
func StartNacosHealthMonitor(configClient config_client.IConfigClient, cfg EnvConfig) {
	healthOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()

			for range ticker.C {
				_, err := configClient.GetConfig(vo.ConfigParam{
					DataId: cfg.NACOS_DATA_ID,
					Group:  cfg.NACOS_GROUP,
				})
				if err != nil {
					UpdateNacosHealth(false, "Nacos connection failed: "+err.Error())
				} else {
					UpdateNacosHealth(true, "")
				}
			}
		}()
	})
}
