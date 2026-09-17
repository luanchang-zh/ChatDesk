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

	"github.com/luanchang-zh/ChatDesk/internal/router"
	v1 "github.com/luanchang-zh/ChatDesk/internal/router/v1"
	"github.com/luanchang-zh/ChatDesk/pkg/logger"
)

func main() {
	if err := logger.InitConsole(); err != nil {
		fmt.Fprintf(os.Stderr, "初始化日志失败: %v\n", err)
		os.Exit(1)
	}

	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	engine := router.NewEngine(v1.NewHealthHandler(), v1.NewConversationHandler())
	server := &http.Server{
		Addr:              addr,
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	runErrCh := make(chan error, 1)
	go func() {
		logger.WithCtx(ctx).Info("HTTP 服务启动中", logger.String("address", addr))
		runErrCh <- server.ListenAndServe()
	}()

	select {
	case err := <-runErrCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.WithCtx(context.Background()).Error("HTTP 服务运行失败", logger.ErrorField("error", err))
			os.Exit(1)
		}
	case <-ctx.Done():
		logger.WithCtx(context.Background()).Warn("收到退出信号，开始关闭 HTTP 服务")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, context.Canceled) {
		logger.WithCtx(context.Background()).Error("关闭 HTTP 服务失败", logger.ErrorField("error", err))
		os.Exit(1)
	}
}
