package main

import (
	"base-core/config"
	"fmt"
	"log/slog"
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

	// 打印环境变量
	fmt.Println("==============================")
	fmt.Println("🚀 基础环境变量加载完成:")
	fmt.Printf("NACOS_SERVER: %s:%d\n", envCfg.SERVER_IP, envCfg.SERVER_PORT)
	fmt.Printf("NAMESPACE:    %s\n", envCfg.NAMESPACE)
	fmt.Printf("DATA_ID:      %s\n", envCfg.DATA_ID)
	fmt.Printf("GROUP:        %s\n", envCfg.GROUP)
	fmt.Println("==============================")

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
	fmt.Println("==============================")
	fmt.Println("📦 Nacos YAML 嵌套配置加载完成:")
	fmt.Printf("AppPort:    %d\n", appCfg.Base.AppPort)
	fmt.Printf("TestMode:   %v\n", appCfg.Base.TestMode)
	fmt.Printf("LarkURL:    %s\n", appCfg.Base.LarkURL)
	fmt.Printf("Domains数:  %d\n", len(appCfg.Domains))
	for _, d := range appCfg.Domains {
		fmt.Printf("  - Domain: %s, SubDomains: %v\n", d.Name, d.SubDomain)
	}
	fmt.Println("==============================")

	// 保持程序运行，以便监听配置变更
	select {}
}
