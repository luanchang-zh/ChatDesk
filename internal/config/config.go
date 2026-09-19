package config

import (
	"os"
	"strings"
)

// Config 进程启动时读入的运行参数。
//
// 约束：
//  1. 只从环境变量取值，不在仓库里写死地址、DSN、密钥；
//  2. 字段都是纯数据，不持有数据库连接；连库是下一轮持久化的事；
//  3. 启动日志可以引用这些字段，但禁止把 PostgresDSN 原文打出去。
type Config struct {
	// HTTPAddr 监听地址，例如 :8080。空时 Load 会填默认值。
	HTTPAddr string
	// PostgresDSN 业务库连接串。空字符串表示还没配库，当前进程仍然可以起 HTTP。
	PostgresDSN string
	// LogLevel zap 级别：debug / info / warn / error。非法值由 logger 回退到 info。
	LogLevel string
}

const (
	envHTTPAddr    = "HTTP_ADDR"
	envPostgresDSN = "POSTGRES_DSN"
	envLogLevel    = "LOG_LEVEL"

	defaultHTTPAddr = ":8080"
	defaultLogLevel = "info"
)

// Load 读取环境变量并填默认值。
// DSN 未配置是合法状态：本轮 handle / service 接口已经能跑，还不需要连 Postgres。
func Load() Config {
	return Config{
		HTTPAddr:    getenv(envHTTPAddr, defaultHTTPAddr),
		PostgresDSN: strings.TrimSpace(os.Getenv(envPostgresDSN)),
		LogLevel:    strings.ToLower(getenv(envLogLevel, defaultLogLevel)),
	}
}

// getenv 读环境变量；全空白视为未配置，改用 fallback。
// 不要把「只含空格」的值当真配置，否则 HTTP_ADDR="  " 会让监听失败。
func getenv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

// PostgresConfigured 只回答「有没有配 DSN」，不尝试拨号。
// 启动日志用这个布尔值，避免把账号密码打进 console。
func (c Config) PostgresConfigured() bool {
	return c.PostgresDSN != ""
}
