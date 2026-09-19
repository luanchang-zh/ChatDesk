package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/luanchang-zh/ChatDesk/internal/config"
	"github.com/luanchang-zh/ChatDesk/internal/database"
	"github.com/luanchang-zh/ChatDesk/internal/repository"
	"github.com/luanchang-zh/ChatDesk/internal/router"
	v1 "github.com/luanchang-zh/ChatDesk/internal/router/v1"
	"github.com/luanchang-zh/ChatDesk/internal/service"
	"github.com/luanchang-zh/ChatDesk/pkg/logger"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	// 1. 环境变量 → Config。DSN 未配、开发身份不合法都会在这里失败。
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// 2. 日志必须在其它业务之前就绪。级别非法时 InitConsole 内部回退 info。
	if err := logger.InitConsole(cfg.LogLevel); err != nil {
		return fmt.Errorf("初始化日志失败: %w", err)
	}

	// 3. 连库、探表，必要时幂等写入开发用户。失败文案不带 DSN。
	startupCtx, startupCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer startupCancel()
	pool, err := database.Open(startupCtx, cfg.PostgresDSN)
	if err != nil {
		return err
	}
	defer pool.Close()
	store := repository.NewConversations(pool)
	if cfg.DevAuthEnabled {
		if err := store.EnsureUser(startupCtx, uuid.MustParse(cfg.DevUserID)); err != nil {
			return errors.New("初始化开发用户失败，请检查数据库权限与结构")
		}
	}
	startupCancel()
	conversations := service.NewConversations(store)
	engine := router.NewEngine(v1.NewHealthHandler(), v1.NewConversationHandler(conversations), cfg.DevUserID)
	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second, // 防止慢客户端占住连接
	}

	// 4. 监听 SIGINT / SIGTERM，Ctrl+C 和容器 stop 走同一条 Shutdown。
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 启动日志只打 DSN 是否已配置，不打印连接串本身。
	logger.WithCtx(ctx).Info("HTTP 服务启动中",
		logger.String("address", cfg.HTTPAddr),
		logger.String("log_level", cfg.LogLevel),
		logger.Bool("postgres_dsn_set", cfg.PostgresConfigured()),
		logger.Bool("dev_auth_enabled", cfg.DevAuthEnabled),
	)

	runErrCh := make(chan error, 1)
	go func() {
		runErrCh <- server.ListenAndServe()
	}()

	select {
	case err := <-runErrCh:
		// ListenAndServe 在 Shutdown 后会返回 ErrServerClosed，这是正常退出。
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("HTTP 服务运行失败: %w", err)
		}
	case <-ctx.Done():
		logger.WithCtx(context.Background()).Warn("收到退出信号，开始关闭 HTTP 服务")
	}

	// 5. 给进行中的请求留收尾时间，超时仍强制退出。
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, context.Canceled) {
		_ = server.Close()
		return fmt.Errorf("关闭 HTTP 服务失败: %w", err)
	}
	return nil
}
