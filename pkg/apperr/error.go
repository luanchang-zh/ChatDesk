package apperr

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/luanchang-zh/ChatDesk/consts"
)

// Error 统一业务错误。
//
//	code    给 HTTP 层映射信封，必须是 consts 里的业务码；
//	message 给信封和 Error() 字符串，走 consts.GetMessage，不要把 SQL 原文放这里；
//	cause   只进日志，不要直接回给浏览器；
//	stack   给 TopFrame 摘要用，console 不打整栈。
type Error struct {
	code    int
	message string
	cause   error
	stack   []uintptr
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	msg := e.message
	if msg == "" {
		msg = consts.GetMessage(e.Code())
	}
	if e.cause != nil {
		return msg + ": " + e.cause.Error()
	}
	return msg
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

func (e *Error) Code() int {
	if e == nil {
		return consts.CodeSuccess
	}
	// 0 在信封里表示成功。错误对象若没带码，只能当成内部错误，不能返回 0。
	if e.code == 0 {
		return consts.CodeInternalError
	}
	return e.code
}

// New 按业务码创建错误，文案走 consts.GetMessage。
// callers(3) 跳过 runtime.Callers → callers → New，让栈顶落在业务调用行。
func New(code int) error {
	return &Error{code: code, message: consts.GetMessage(code), stack: callers(3)}
}

// Wrap 给底层错误补上业务码。err 为 nil 时不包装，避免出现「成功的 Wrap」。
// msg 为空时同样走 consts，避免各处手写不一致的中文。
func Wrap(err error, code int, msg string) error {
	if err == nil {
		return nil
	}
	if msg == "" {
		msg = consts.GetMessage(code)
	}
	return &Error{code: code, message: msg, cause: err, stack: callers(3)}
}

// Code 从错误链里取出业务码。
// 不是本包 Error 的（驱动、context.Canceled、fmt.Errorf），一律当内部错误，
// 避免把驱动原文或「context canceled」当成业务码回给前端。
func Code(err error) int {
	if err == nil {
		return consts.CodeSuccess
	}
	var app *Error
	if errors.As(err, &app) {
		return app.Code()
	}
	return consts.CodeInternalError
}

// NewFromPanic Recovery 专用。skip 比 New 多一帧，才能跳过 recover 包装函数本身。
func NewFromPanic(recoverVal any) error {
	return &Error{
		code:    consts.CodeInternalError,
		message: consts.GetMessage(consts.CodeInternalError),
		cause:   fmt.Errorf("panic: %v", recoverVal),
		stack:   callers(4),
	}
}

// TopFrame 返回最靠近业务的一帧，方便日志摘要，不必把整栈塞进 console。
func TopFrame(err error) string {
	var app *Error
	if !errors.As(err, &app) || len(app.stack) == 0 {
		return ""
	}
	frames := runtime.CallersFrames(app.stack)
	frame, ok := frames.Next()
	if !ok {
		return ""
	}
	fn := frame.Function
	if idx := strings.LastIndex(fn, "/"); idx >= 0 {
		fn = fn[idx+1:]
	}
	if frame.File == "" {
		return fn
	}
	return fmt.Sprintf("%s %s:%d", fn, filepath.Base(frame.File), frame.Line)
}

// callers skip 从 0 起算：0=Callers 自己，1=callers，2=New/Wrap，3=业务调用点。
func callers(skip int) []uintptr {
	pcs := make([]uintptr, 16)
	n := runtime.Callers(skip, pcs)
	if n == 0 {
		return nil
	}
	out := make([]uintptr, n)
	copy(out, pcs[:n])
	return out
}
