// types.go
package config

import (
	"log/slog"
	"sync"
	"time"

	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
)

// EnvConfig 结构体定义基础环境变量配置（用于连接 Nacos）
type EnvConfig struct {
	SERVER_IP   string
	SERVER_PORT uint64
	NAMESPACE   string
	DATA_ID     string
	GROUP       string
	NACOSUSER   string
	PASSWORD    string
	LOG_LEVEL   slog.Level
}

// Debouncer 泛型防抖器
type Debouncer[T any] struct {
	ch     chan T
	delay  time.Duration
	onFire func(T)
	stop   chan struct{}
	done   chan struct{}
	once   sync.Once
}

// getNacosClient 是线程安全的 nacosClient 访问辅助函数

func getNacosClient() config_client.IConfigClient {
	nacosMu.RLock()
	defer nacosMu.RUnlock()
	return nacosClient
}
