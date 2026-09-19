package config

import (
	"fmt"
	"net"
	"net/netip"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// Config 进程启动时读入的运行参数。
//
// 约束：
//  1. 只从环境变量取值，不在仓库里写死地址、DSN、密钥；
//  2. 字段都是纯数据，不持有数据库连接；
//  3. 启动日志可以引用这些字段，但禁止把 PostgresDSN 原文打出去。
type Config struct {
	// HTTPAddr 监听地址，例如 :8080。空时 Load 会填默认值。
	HTTPAddr string
	// PostgresDSN 业务库连接串，服务启动前必须配置。
	PostgresDSN string
	// LogLevel zap 级别：debug / info / warn / error。非法值由 logger 回退到 info。
	LogLevel string
	// DevAuthEnabled 是否显式开启本机开发身份。默认 false，业务接口未认证。
	DevAuthEnabled bool
	// DevUserID 开发身份的用户 UUID。仅在 DevAuthEnabled 为 true 时有值。
	DevUserID string
}

const (
	envHTTPAddr    = "HTTP_ADDR"
	envPostgresDSN = "POSTGRES_DSN"
	envLogLevel    = "LOG_LEVEL"

	defaultHTTPAddr = ":8080"
	defaultLogLevel = "info"
)

// Load 读取环境变量并填默认值。
// 开发身份开启时默认监听回环；HTTP_ADDR 若写成 :8080 或外网 IP 会直接失败。
func Load() (Config, error) {
	devAuth, err := strconv.ParseBool(getenv("DEV_AUTH_ENABLED", "false"))
	if err != nil {
		return Config{}, fmt.Errorf("DEV_AUTH_ENABLED 必须是布尔值")
	}
	defaultAddr := defaultHTTPAddr
	if devAuth {
		defaultAddr = "127.0.0.1:8080"
	}
	cfg := Config{
		HTTPAddr:       getenv(envHTTPAddr, defaultAddr),
		PostgresDSN:    strings.TrimSpace(os.Getenv(envPostgresDSN)),
		LogLevel:       strings.ToLower(getenv(envLogLevel, defaultLogLevel)),
		DevAuthEnabled: devAuth,
	}
	if cfg.PostgresDSN == "" {
		return Config{}, fmt.Errorf("POSTGRES_DSN 未配置")
	}
	host, port, err := net.SplitHostPort(cfg.HTTPAddr)
	if err != nil {
		return Config{}, fmt.Errorf("HTTP_ADDR 必须包含监听地址和端口")
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 0 || portNumber > 65535 {
		return Config{}, fmt.Errorf("HTTP_ADDR 端口无效")
	}
	if devAuth {
		addr, err := netip.ParseAddr(host)
		if err != nil || !addr.IsLoopback() || addr.Zone() != "" {
			return Config{}, fmt.Errorf("开发身份仅允许监听回环 IP，例如 127.0.0.1:8080")
		}
		id, err := uuid.Parse(strings.TrimSpace(os.Getenv("DEV_USER_ID")))
		if err != nil || id == uuid.Nil {
			return Config{}, fmt.Errorf("开发身份需要有效的非零 DEV_USER_ID UUID")
		}
		cfg.DevUserID = id.String()
	}
	return cfg, nil
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
