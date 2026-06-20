package config

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"gopkg.in/yaml.v2"
)

var (
	nacosClient config_client.IConfigClient
	nacosMu     sync.RWMutex
)

// InitNacosClient 根据环境配置初始化 Nacos 配置中心客户端。
// 在调用 LoadNacosConfig 或 FetchAppConfigFromNacos 前需先调用此函数。
func InitNacosClient(cfg EnvConfig) error {
	serverConfig := []constant.ServerConfig{
		{IpAddr: cfg.NACOS_SERVER_IP, Port: cfg.NACOS_SERVER_PORT},
	}

	clientConfig := constant.ClientConfig{
		NamespaceId: cfg.NACOS_NAMESPACE,
		TimeoutMs:   1000,
		LogDir:      "./logs",
		CacheDir:    "./cache",
		LogLevel:    "debug",
		Username:    cfg.NACOS_USER,
		Password:    cfg.NACOS_PASSWORD,
	}

	client, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": serverConfig,
		"clientConfig":  clientConfig,
	})
	if err != nil {
		return err
	}
	slog.Info("nacos client 初始化成功")

	nacosMu.Lock()
	nacosClient = client
	nacosMu.Unlock()
	return nil
}

func FetchAppConfigFromNacos[T any](cfg EnvConfig, dest *T) error {
	client := getNacosClient()
	if client == nil {
		return fmt.Errorf("nacos client 未初始化")
	}
	content, err := client.GetConfig(vo.ConfigParam{
		DataId: cfg.NACOS_DATA_ID,
		Group:  cfg.NACOS_GROUP,
	})
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal([]byte(content), dest); err != nil {
		return err
	}
	return nil
}
