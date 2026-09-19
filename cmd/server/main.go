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

	"github.com/luanchang-zh/ChatDesk/internal/config"
	"github.com/luanchang-zh/ChatDesk/internal/router"
	v1 "github.com/luanchang-zh/ChatDesk/internal/router/v1"
	"github.com/luanchang-zh/ChatDesk/internal/service"
	"github.com/luanchang-zh/ChatDesk/pkg/logger"
)

// main 只做进程级编排，不写业务：
//  1. 读配置；
//  2. 初始化日志（失败时 logger 还不可用，只能写 stderr）；
//  3. 注入领域服务（当前是占位实现）；
//  4. 监听信号，先停 HTTP 入口再退出。
func main() {
	// 1. 环境变量 → Config。DSN 未配也能启动，方便先调 HTTP 信封。
	cfg := config.Load()

	// 2. 日志必须在其它业务之前就绪。级别非法时 InitConsole 内部回退 info。
	if err := logger.InitConsole(cfg.LogLevel); err != nil {
		fmt.Fprintf(os.Stderr, "初始化日志失败: %v\n", err)
		os.Exit(1)
	}

	// 3. 组装 HTTP。会话 handle 只拿接口，下一轮换成 sqlc 实现时改这一行即可。
	conversations := service.UnavailableConversations{}
	engine := router.NewEngine(v1.NewHealthHandler(), v1.NewConversationHandler(conversations))
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
	)

	runErrCh := make(chan error, 1)
	go func() {
		runErrCh <- server.ListenAndServe()
	}()

	select {
	case err := <-runErrCh:
		// ListenAndServe 在 Shutdown 后会返回 ErrServerClosed，这是正常退出。
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.WithCtx(context.Background()).Error("HTTP 服务运行失败", logger.ErrorField("error", err))
			os.Exit(1)
		}
	case <-ctx.Done():
		logger.WithCtx(context.Background()).Warn("收到退出信号，开始关闭 HTTP 服务")
	}

	// 5. 给进行中的请求留收尾时间，超时仍强制退出。
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, context.Canceled) {
		logger.WithCtx(context.Background()).Error("关闭 HTTP 服务失败", logger.ErrorField("error", err))
		os.Exit(1)
	}
}
