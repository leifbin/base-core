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

func InitNacosClient(cfg EnvConfig) error {
	serverConfig := []constant.ServerConfig{
		{IpAddr: cfg.SERVER_IP, Port: cfg.SERVER_PORT},
	}

	clientConfig := constant.ClientConfig{
		NamespaceId: cfg.NAMESPACE,
		TimeoutMs:   1000,
		LogDir:      "./logs",
		CacheDir:    "./cache",
		LogLevel:    "debug",
		Username:    cfg.NACOSUSER,
		Password:    cfg.PASSWORD,
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
		DataId: cfg.DATA_ID,
		Group:  cfg.GROUP,
	})
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal([]byte(content), dest); err != nil {
		return err
	}
	return nil
}
