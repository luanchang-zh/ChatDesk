package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/luanchang-zh/ChatDesk/internal/database"
)

// 独立迁移入口。HTTP 进程不会调用这里，避免启动时隐式改库。
func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) != 1 || (args[0] != "up" && args[0] != "down" && args[0] != "status") {
		return fmt.Errorf("用法：go run ./cmd/migrate <up|down|status>；连接串通过 POSTGRES_DSN 配置")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	// DSN 只走环境变量，不接受命令行参数，避免出现在进程列表里。
	provider, err := database.NewMigrator(ctx, os.Getenv("POSTGRES_DSN"))
	if err != nil {
		return err
	}
	defer provider.Close()
	switch args[0] {
	case "up":
		_, err = provider.Up(ctx)
	case "down":
		// down 只回滚最近一个版本；当前首个迁移会删表，只能对可丢弃的库执行。
		_, err = provider.Down(ctx)
	case "status":
		statuses, statusesErr := provider.Status(ctx)
		if statusesErr != nil {
			return fmt.Errorf("读取数据库迁移状态失败")
		}
		for _, status := range statuses {
			fmt.Printf("%d\t%s\n", status.Source.Version, status.State)
		}
	}
	if err != nil {
		return fmt.Errorf("数据库迁移 %s 失败，请检查数据库状态及迁移文件", args[0])
	}
	return nil
}
