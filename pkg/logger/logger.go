package logger

import (
	"context"
	"os"
	"time"

	"github.com/luanchang-zh/ChatDesk/pkg/ctxmeta"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var global = zap.NewNop()

// L 返回全局 logger。未 Init 时是 Nop，调用不会崩。
func L() *zap.Logger {
	return global
}

// InitConsole 初始化控制台 logger。
// 控制台 encoder，不用 JSON；字段仍然走 zap.Field。
func InitConsole() error {
	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     zapcore.TimeEncoderOfLayout(time.RFC3339Nano),
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		zapcore.AddSync(os.Stdout),
		zap.NewAtomicLevelAt(zap.InfoLevel),
	)
	global = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	zap.ReplaceGlobals(global)
	return nil
}

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

// CtxLogger 带请求上下文的 logger。习惯写法：logger.WithCtx(ctx).Error(...)
type CtxLogger struct {
	ctx context.Context
}

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

func String(key, value string) zap.Field {
	return zap.String(key, value)
}

func Int(key string, value int) zap.Field {
	return zap.Int(key, value)
}

func Duration(key string, value time.Duration) zap.Field {
	return zap.Duration(key, value)
}

func Any(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}

func ErrorField(key string, err error) zap.Field {
	return zap.Error(err)
}
