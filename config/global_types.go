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
	NACOS_SERVER_IP   string
	NACOS_SERVER_PORT uint64
	NACOS_NAMESPACE   string
	NACOS_DATA_ID     string
	NACOS_GROUP       string
	NACOS_USER        string
	NACOS_PASSWORD    string
	LOG_LEVEL         slog.Level
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
	onError    func(string, error)
}

// watcherData 防抖器内部使用的数据结构
type watcherData[T any] struct {
	config *T
	diffs  []ConfigDiff
}

// OnError 注册配置解析失败时的回调
func (w *Watcher[T]) OnError(handler func(dataId string, err error)) *Watcher[T] {
	w.onError = handler
	return w
}
