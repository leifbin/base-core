package config

import (
	"sync"
	"time"

	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

// NacosHealth 全局变量，记录 Nacos 配置中心的健康检查状态。
var NacosHealth struct {
	Healthy   bool
	LastCheck time.Time
	Error     string
	Mutex     sync.RWMutex
}

// UpdateNacosHealth 更新 Nacos 健康状态（线程安全）。
func UpdateNacosHealth(healthy bool, errMsg string) {
	NacosHealth.Mutex.Lock()
	defer NacosHealth.Mutex.Unlock()

	NacosHealth.Healthy = healthy
	NacosHealth.LastCheck = time.Now()
	NacosHealth.Error = errMsg
}

// StartNacosHealthMonitor 启动后台 goroutine，定期（30秒）检查 Nacos 连接状态。
// 检查结果会更新到 NacosHealth 全局变量。
func StartNacosHealthMonitor(configClient config_client.IConfigClient, cfg EnvConfig, stop <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				_, err := configClient.GetConfig(vo.ConfigParam{
					DataId: cfg.DATA_ID,
					Group:  cfg.GROUP,
				})
				if err != nil {
					UpdateNacosHealth(false, "Nacos connection failed: "+err.Error())
				} else {
					UpdateNacosHealth(true, "")
				}
			case <-stop:
				return
			}
		}
	}()
}
