package apperr

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/luanchang-zh/ChatDesk/consts"
)

// Error 统一业务错误。业务码给 HTTP 层映射，cause 只进日志。
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
	if e.code == 0 {
		return consts.CodeInternalError
	}
	return e.code
}

func New(code int) error {
	return &Error{code: code, message: consts.GetMessage(code), stack: callers(3)}
}

func Wrap(err error, code int, msg string) error {
	if err == nil {
		return nil
	}
	if msg == "" {
		msg = consts.GetMessage(code)
	}
	return &Error{code: code, message: msg, cause: err, stack: callers(3)}
}

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

func NewFromPanic(recoverVal any) error {
	return &Error{
		code:    consts.CodeInternalError,
		message: consts.GetMessage(consts.CodeInternalError),
		cause:   fmt.Errorf("panic: %v", recoverVal),
		stack:   callers(4),
	}
}

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
