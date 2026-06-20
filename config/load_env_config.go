package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
)

// LoadEnvConfig 从环境变量加载基础配置（包含 Nacos 连接信息）
func LoadEnvConfig() EnvConfig {
	return EnvConfig{
		NACOS_SERVER_IP:   getEnv("NACOS_SERVER_IP", "127.0.0.1"),
		NACOS_SERVER_PORT: getEnvAsUint64("NACOS_SERVER_PORT", 8848),
		NACOS_NAMESPACE:   getEnv("NACOS_NAMESPACE", ""),
		NACOS_DATA_ID:     getEnv("NACOS_DATA_ID", "base-core.yaml"),
		NACOS_GROUP:       getEnv("NACOS_GROUP", "DEFAULT_GROUP"),
		NACOS_USER:        getEnv("NACOS_USER", "nacos"),
		NACOS_PASSWORD:    getEnv("NACOS_PASSWORD", "nacos"),
		LOG_LEVEL:         getLogLevel(),
	}
}

// getEnv 获取字符串类型的环境变量，如果未设置则返回默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsUint64 获取 uint64 类型的环境变量，如果未设置则返回默认值
func getEnvAsUint64(key string, defaultValue uint64) uint64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseUint(valueStr, 10, 64)
	if err != nil {
		return defaultValue
	}
	return value
}

// getLogLevel 获取日志级别
func getLogLevel() slog.Level {
	levelStr := os.Getenv("LOG_LEVEL") // 读取环境变量
	if levelStr == "" {
		return slog.LevelInfo // 如果没有设置环境变量，默认返回 INFO，不报错
	}
	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		tmpLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
		tmpLogger.Error("⚠️ 无效的 LOG_LEVEL, 默认使用 INFO",
			"invalid_level", levelStr,
			"suggestion", "DEBUG, INFO, WARN, ERROR",
		)
		return slog.LevelInfo // 默认级别
	}
}
