# Base-Core 配置加载器

本项目作为一个基础库，用于从环境变量加载 Nacos 连接信息，并从 Nacos 获取 YAML 格式的业务配置。

## 功能特性

- **多配置监听** — 支持同时监听多个 Nacos DataId，每个独立管理
- **泛型支持** — 可自定义任意 Go 结构体接收配置
- **变更差异比对** — 配置变更时自动递归对比结构体差异
- **防抖机制** — 配置频繁变更时 20 秒防抖合并
- **健康检查** — 每 30 秒检查 Nacos 连接状态
- **优雅关闭** — 支持信号处理和资源清理

## 支持的环境变量

### Nacos 连接配置
| 环境变量 | 说明 | 默认值 |
| :--- | :--- | :--- |
| `NACOS_SERVER_IP` | Nacos 服务器地址 | `127.0.0.1` |
| `NACOS_SERVER_PORT` | Nacos 服务器端口 | `8848` |
| `NACOS_NAMESPACE` | Nacos 命名空间 (Namespace ID) | `""` |
| `NACOS_DATA_ID` | Nacos Data ID | `base-core.yaml` |
| `NACOS_GROUP` | Nacos 分组 (Group) | `DEFAULT_GROUP` |
| `NACOS_USER` | Nacos 用户名 | `nacos` |
| `NACOS_PASSWORD` | Nacos 密码 | `nacos` |

### 系统配置
| 环境变量 | 说明 | 默认值 |
| :--- | :--- | :--- |
| `LOG_LEVEL` | 日志级别 (DEBUG, INFO, WARN, ERROR) | `INFO` |

## 使用示例 (单配置)

```go
type AppConfig struct {
    Base struct {
        Port int `yaml:"app_port"`
    } `yaml:"base"`
    Domains []struct {
        Name string `yaml:"name"`
    } `yaml:"domains"`
}

var cfg AppConfig
watcher := config.NewWatcherAppConfig
cleanup, err := watcher.Load(&cfg, func(newCfg *AppConfig, diffs []config.ConfigDiff) {
    for _, d := range diffs {
        fmt.Printf("变更: %s, 旧: %v, 新: %v\n", d.Path, d.OldValue, d.NewValue)
    }
})
defer cleanup()
```

## 使用示例 (多配置)

```go
type RedisConfig struct {
    Redis struct {
        Mode     string `yaml:"mode"`
        Addr     string `yaml:"addr"`
        Password string `yaml:"password"`
        PoolSize int    `yaml:"pool_size"`
    } `yaml:"redis"`
}

domainEnv := envCfg
domainEnv.DATA_ID = "domain.yaml"
domainWatcher := config.NewWatcherAppConfig
cleanup1, _ := domainWatcher.Load(&appCfg, onDomainChange)
defer cleanup1()

redisEnv := envCfg
redisEnv.DATA_ID = "redis.yaml"
redisWatcher := config.NewWatcherRedisConfig
cleanup2, _ := redisWatcher.Load(&redisCfg, onRedisChange)
defer cleanup2()
```

## 启动示例 (Shell)

```bash
export NACOS_SERVER_IP="your-nacos-server"
export NACOS_SERVER_PORT=8848
export NACOS_NAMESPACE="test"
export NACOS_DATA_ID="domain.yaml"
export NACOS_GROUP="DEFAULT_GROUP"
export NACOS_USER="nacos"
export NACOS_PASSWORD="nacos"
export LOG_LEVEL="DEBUG"

go run main.go
```

## 启动示例 (PowerShell)

```powershell
$env:NACOS_SERVER_IP = "your-nacos-server"
$env:NACOS_SERVER_PORT = "8848"
$env:NACOS_NAMESPACE = "devops"
$env:NACOS_DATA_ID = "domain.yaml"
$env:NACOS_GROUP = "DEFAULT_GROUP"
$env:NACOS_USER = "nacos"
$env:NACOS_PASSWORD = "nacos"
$env:LOG_LEVEL = "DEBUG"

go run main.go
```

## 编译参数

```
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -buildvcs=false -o app
```

| 参数 | 类型 | 含义 | 备注 |
| :--- | :--- | :--- | :--- |
| `CGO_ENABLED=0` | 环境变量 | 禁用 CGO，纯 Go 静态链接 | 适合容器/跨平台部署 |
| `GOOS=linux` | 环境变量 | 目标系统 Linux | 交叉编译 |
| `GOARCH=amd64` | 环境变量 | 目标架构 x86_64 | 默认当前机器架构 |
| `-ldflags="-s -w"` | 编译参数 | 去掉符号表和调试信息 | 减小二进制体积 |
| `-buildvcs=false` | 编译参数 | 不记录版本控制信息 | |
| `-o app` | 编译参数 | 输出文件名 | |