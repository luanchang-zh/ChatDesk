package logger

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/luanchang-zh/ChatDesk/pkg/ctxmeta"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// global 未初始化时是 Nop：业务代码调用 Info/Error 不会 panic。
// InitConsole 成功后替换成真正的 console logger。
var global = zap.NewNop()

// L 返回底层 zap.Logger。一般业务请用 WithCtx / Error(ctx, ...)，不要直接拼 zap 字段名。
func L() *zap.Logger {
	return global
}

// InitConsole 初始化控制台 logger。必须在其它业务之前调用。
//
// 约定：
//   - 编码用 console，不用 JSON，本地好读；字段仍然是结构化 zap.Field；
//   - level 非法或空串时回退 info，避免 LOG_LEVEL 写错导致进程起不来；
//   - AddCallerSkip(1) 跳过本包封装，日志里的 caller 指向业务行号。
func InitConsole(level string) error {
	atomicLevel := zap.NewAtomicLevelAt(zap.InfoLevel)
	if err := atomicLevel.UnmarshalText([]byte(strings.ToLower(strings.TrimSpace(level)))); err != nil {
		atomicLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     zapcore.TimeEncoderOfLayout(time.RFC3339Nano),
		EncodeDuration: zapcore.MillisDurationEncoder, // 耗时用毫秒，访问日志更好读
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		zapcore.AddSync(os.Stdout),
		atomicLevel,
	)
	global = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	zap.ReplaceGlobals(global)
	return nil
}

// appendContextFields 把 ctx 里的链路字段拼进本条日志。
// ctx 为 nil 时不加字段，避免后台 goroutine 忘传 context 直接崩。
// 空字段不输出，避免每条日志都带 user_uuid=""。
func appendContextFields(ctx context.Context, fields []zap.Field) []zap.Field {
	if ctx == nil {
		return fields
	}
	if traceID := ctxmeta.TraceID(ctx); traceID != "" {
		fields = append(fields, zap.String(ctxmeta.KeyTraceID, traceID))
	}
	if userUUID := ctxmeta.UserUUID(ctx); userUUID != "" {
		fields = append(fields, zap.String(ctxmeta.KeyUserUUID, userUUID))
	}
	if clientIP := ctxmeta.ClientIP(ctx); clientIP != "" {
		fields = append(fields, zap.String(ctxmeta.KeyClientIP, clientIP))
	}
	return fields
}

// Info / Warn / Error / Debug 是「直接传 ctx」的写法。
// 更常见的是 logger.WithCtx(ctx).Error(...)，两套都保留，调用习惯对齐即可。

func Info(ctx context.Context, msg string, fields ...zap.Field) {
	global.Info(msg, appendContextFields(ctx, fields)...)
}

func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	global.Warn(msg, appendContextFields(ctx, fields)...)
}

func Error(ctx context.Context, msg string, fields ...zap.Field) {
	global.Error(msg, appendContextFields(ctx, fields)...)
}

func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	global.Debug(msg, appendContextFields(ctx, fields)...)
}

// CtxLogger 带请求上下文的 logger。
// 习惯写法：logger.WithCtx(ctx).Error("...", logger.ErrorField("error", err))
type CtxLogger struct {
	ctx context.Context
}

// WithCtx 包一层 context。nil 会换成 Background，保证永远能打。
func WithCtx(ctx context.Context) CtxLogger {
	if ctx == nil {
		ctx = context.Background()
	}
	return CtxLogger{ctx: ctx}
}

func (l CtxLogger) Info(msg string, fields ...zap.Field) {
	global.Info(msg, appendContextFields(l.ctx, fields)...)
}

func (l CtxLogger) Warn(msg string, fields ...zap.Field) {
	global.Warn(msg, appendContextFields(l.ctx, fields)...)
}

func (l CtxLogger) Error(msg string, fields ...zap.Field) {
	global.Error(msg, appendContextFields(l.ctx, fields)...)
}

func (l CtxLogger) Debug(msg string, fields ...zap.Field) {
	global.Debug(msg, appendContextFields(l.ctx, fields)...)
}

// 下面是字段工厂。业务代码不要直接 import zap，避免各处 API 不一致。

func String(key, value string) zap.Field {
	return zap.String(key, value)
}

func Int(key string, value int) zap.Field {
	return zap.Int(key, value)
}

func Bool(key string, value bool) zap.Field {
	return zap.Bool(key, value)
}

func Duration(key string, value time.Duration) zap.Field {
	return zap.Duration(key, value)
}

func Any(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}

// ErrorField 第一个参数是为了和习惯调用 logger.ErrorField("error", err) 对齐。
// zap.Error 本身会写成 error 键，key 目前不参与输出。
func ErrorField(key string, err error) zap.Field {
	return zap.Error(err)
}
