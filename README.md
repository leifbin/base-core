# Base-Core 配置加载器

本项目作为一个基础库，用于从环境变量加载 Nacos 连接信息，并从 Nacos 获取 YAML 格式的业务配置。

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

## 使用示例 (作为 Core 包)

在你的项目中定义自定义的配置结构体：

```go
type AppConfig struct {
    Base struct {
        Port int `yaml:"app_port"`
    } `yaml:"base"`
    Domains []struct {
        Name string `yaml:"name"`
    } `yaml:"domains"`
}

// 加载配置
var cfg AppConfig
err := config.LoadNacosConfig(envCfg, &cfg, func(newCfg *AppConfig, diffs []config.ConfigDiff) {
    // 处理配置变更
    for _, d := range diffs {
        fmt.Printf("变更: %s, 旧: %v, 新: %v\n", d.Path, d.OldValue, d.NewValue)
    }
})
```

## 启动示例 (Shell)

```bash
export NACOS_SERVER_IP="afdf02d7f82004e9c835dd9a6ac31f6f-aff024302e5247be.elb.ap-east-1.amazonaws.com"
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
$env:NACOS_SERVER_IP = "afdf02d7f82004e9c835dd9a6ac31f6f-aff024302e5247be.elb.ap-east-1.amazonaws.com"
$env:NACOS_SERVER_PORT = "8848"
$env:NACOS_NAMESPACE = "devops"
$env:NACOS_DATA_ID = "domain.yaml"
$env:NACOS_GROUP = "DEFAULT_GROUP"
$env:NACOS_USER = "nacos"
$env:NACOS_PASSWORD = "nacos"
$env:LOG_LEVEL = "DEBUG"

go run main.go
```

```
编译
参数 / 环境变量	类型	含义	备注
CGO_ENABLED=0	环境变量	禁用 CGO，生成纯 Go 静态链接二进制	避免依赖系统 C 库，更适合容器或跨平台部署
GOOS=linux	环境变量	指定目标操作系统为 Linux	交叉编译时使用，比如在 Mac 或 Windows 上生成 Linux 二进制
GOARCH=amd64	环境变量	指定目标 CPU 架构为 x86_64	如果不加，默认使用当前机器架构
-ldflags="-s -w"	编译参数	去掉符号表(-s)和调试信息(-w)	生成更小的二进制文件
-buildvcs=false	编译参数	不在二进制中记录版本控制信息	默认 true，会写 commit hash 等信息到二进制
-o app	编译参数	指定输出二进制文件名为 app	默认输出是当前目录下的源文件名

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w"  -buildvcs=false -o app 
```
