package main

import (
	"base-core/config"
	"fmt"
	"log/slog"
	"os"
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

func main() {
	// 1. 加载环境变量（包含 Nacos 连接信息）
	envCfg := config.LoadEnvConfig()
	//初始化日志
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: envCfg.LOG_LEVEL, // 通过 config 获取日志级别
	}))
	slog.SetDefault(logger)
	// 打印环境变量
	slog.Debug("🚀 基础环境变量加载完成:")
	slog.Debug("NACOS_SERVER", "NACOS_SERVER", envCfg.SERVER_IP)
	slog.Debug("NACOS_PORT", "NACOS_PORT", envCfg.SERVER_PORT)
	slog.Debug("NAMESPACE", "NAMESPACE", envCfg.NAMESPACE)
	slog.Debug("DATA_ID", "DATA_ID", envCfg.DATA_ID)
	slog.Debug("GROUP", "GROUP", envCfg.GROUP)

	// 2. 初始化 Nacos 客户端
	err := config.InitNacosClient(envCfg)
	if err != nil {
		slog.Error("初始化 Nacos 客户端失败", "err", err)
		return
	}

	//// 3. 从 Nacos 获取并监听 YAML 配置
	// 这里演示如何使用自定义的 AppConfig 结构，并接收变更列表
	var appCfg AppConfig
	err = config.LoadNocosConfig(envCfg, &appCfg, func(newCfg *AppConfig, diffs []config.ConfigDiff) {
		slog.Info("🔔 Nacos 配置发生变更", "diff_count", len(diffs))
		for _, d := range diffs {
			slog.Info(fmt.Sprintf("变更详情: 路径=%s, 类型=%s", d.Path, d.Type))
		}
	})

	if err != nil {
		slog.Error("加载 Nacos 配置失败", "err", err)
		return
	}

	// 4. 打印从 Nacos 获取的 YAML 配置内容
	slog.Debug("📦 Nacos YAML 嵌套配置加载完成:")
	slog.Debug("AppPort", appCfg.Base.AppPort)
	slog.Debug("TestMode", appCfg.Base.TestMode)
	slog.Debug("LarkURL", appCfg.Base.LarkURL)
	slog.Debug("Domains数:", len(appCfg.Domains))
	for _, d := range appCfg.Domains {
		slog.Debug("Domain",
			slog.String("domain", d.Name),
			slog.Any("sub_domains", d.SubDomain),
		)
	}

	// 保持程序运行，以便监听配置变更
	select {}
}
