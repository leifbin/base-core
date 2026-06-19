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

// Watcher 管理单个 DataId 的 Nacos 配置加载和监听实例。
// 每个 DataId 对应一个独立的 Watcher，各自维护配置快照和防抖器。
type Watcher[T any] struct {
	envCfg     EnvConfig
	lastConfig interface{}
	mu         sync.RWMutex
	debouncer  *Debouncer[*watcherData[T]]
	stopCh     chan struct{}
	cleanup    func()
}

// watcherData 防抖器内部使用的数据结构
type watcherData[T any] struct {
	config *T
	diffs  []ConfigDiff
}
