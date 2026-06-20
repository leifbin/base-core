package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/leifbin/base-core/config"
)

// DomainConfig 子域名配置
type DomainConfig struct {
	Name      string   `yaml:"name"`
	SubDomain []string `yaml:"sub_domain"`
}

// BaseSettings 基础配置项
type BaseSettings struct {
	AppPort                 uint64 `yaml:"app_port"`
	TestMode                bool   `yaml:"test_mode"`
	SubConcurrencyStopLimit int    `yaml:"subConcurrencyStopLimit"`
	TaskInterval            int    `yaml:"task_interval"`
	MainConcurrencyLimit    int    `yaml:"main_concurrency_limit"`
	SubConcurrencyLimit     int    `yaml:"sub_concurrency_limit"`
	LarkURL                 string `yaml:"larkurl"`
}

// AppConfig 这是具体项目自定义的嵌套配置结构
type AppConfig struct {
	Base    BaseSettings   `yaml:"base"`
	Domains []DomainConfig `yaml:"domains"`
}

type RedisConfig struct {
	Redis RedisSettings `yaml:"redis"`
}

type RedisSettings struct {
	Mode         string   `yaml:"mode"`
	Addr         string   `yaml:"addr"`
	Password     string   `yaml:"password"`
	ClusterAddrs []string `yaml:"cluster_addrs"`

	PoolSize     int `yaml:"pool_size"`
	MinIdleConns int `yaml:"min_idle_conns"`
	MaxIdleConns int `yaml:"max_idle_conns"`
	PoolTimeout  int `yaml:"pool_timeout"`

	DialTimeout  int `yaml:"dial_timeout"`
	ReadTimeout  int `yaml:"read_timeout"`
	WriteTimeout int `yaml:"write_timeout"`

	MaxRetries      int `yaml:"max_retries"`
	MinRetryBackoff int `yaml:"min_retry_backoff"`
	MaxRetryBackoff int `yaml:"max_retry_backoff"`

	MaxRedirects   int  `yaml:"max_redirects"`
	ReadOnly       bool `yaml:"read_only"`
	RouteByLatency bool `yaml:"route_by_latency"`
	RouteRandomly  bool `yaml:"route_randomly"`
}

func main() {
	envCfg := config.LoadEnvConfig()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: envCfg.LOG_LEVEL,
	}))
	slog.SetDefault(logger)

	// ====== 加载第1个配置: domain.yaml ======
	var appCfg AppConfig
	watcher1 := config.NewWatcher[AppConfig](envCfg)
	cleanup1, err := watcher1.Load(&appCfg, func(newCfg *AppConfig, diffs []config.ConfigDiff) {
		slog.Info("🔔 domain.yaml 配置发生变更", "diff_count", len(diffs))
		for _, d := range diffs {
			slog.Info("变更详情", "path", d.Path, "type", d.Type)
		}
	})
	if err != nil {
		slog.Error("加载 domain.yaml 失败", "err", err)
		return
	}
	defer cleanup1()

	slog.Debug("📦 domain.yaml 加载完成:")
	slog.Debug("AppPort", "port", appCfg.Base.AppPort)
	slog.Debug("TestMode", "testMode", appCfg.Base.TestMode)
	slog.Debug("Domains数", "count", len(appCfg.Domains))

	// ====== 加载第2个配置: redis.yaml ======
	redisEnv := envCfg
	redisEnv.NACOS_DATA_ID = "redis.yaml" // ← 覆盖 DataId

	var redisCfg RedisConfig
	watcher2 := config.NewWatcher[RedisConfig](redisEnv)
	cleanup2, err := watcher2.Load(&redisCfg, func(newCfg *RedisConfig, diffs []config.ConfigDiff) {
		slog.Info("🔔 redis.yaml 配置发生变更", "diff_count", len(diffs))
		for _, d := range diffs {
			slog.Info("变更详情", "path", d.Path, "type", d.Type)
		}
	})
	if err != nil {
		slog.Error("加载 redis.yaml 失败", "err", err)
		return
	}
	defer cleanup2()

	slog.Debug("📦 redis.yaml 加载完成:")
	slog.Debug("Redis", "mode", redisCfg.Redis.Mode, "addr", redisCfg.Redis.Addr)
	slog.Debug("Pool", "poolSize", redisCfg.Redis.PoolSize, "minIdle", redisCfg.Redis.MinIdleConns, "maxIdle", redisCfg.Redis.MaxIdleConns)

	// 等待退出信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	slog.Info("收到退出信号，正在关闭...", "signal", sig)
}
