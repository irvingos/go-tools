package middleware

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/irvingos/go-tools/errorx"
	"github.com/irvingos/go-tools/logx"
	"github.com/irvingos/go-tools/resp"
)

const maxStackBytes = 64 << 10 // 64KB

func RecoveryMiddleware(hideHeaders []string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ecw := &errCapturingWriter{ResponseWriter: ctx.Writer}
		ctx.Writer = ecw
		defer func() {
			if r := recover(); r != nil {
				if isBrokenPipe(r, ctx.Request) || isBrokenPipe(ecw.LastErr(), ctx.Request) {
					logx.WithContext(ctx).
						WithField(logx.FieldEvent, "panic_recovered_broken_pipe").
						Warn()
					return
				}

				stack := ""
				b := debug.Stack()
				if len(b) > maxStackBytes {
					b = b[:maxStackBytes]
				}
				stack = string(b)

				fn, file, line := topFrame()

				logx.WithContext(ctx).
					WithField(logx.FieldEvent, "panic_recovered").
					WithField(logx.FieldStack, stack).
					WithField(logx.FieldHeaders, filterHeaders(ctx.Request.Header, hideHeaders)).
					WithField(logx.FieldFile, fmt.Sprintf("%s:%d", file, line)).
					WithField(logx.FieldFunction, fn).
					WithField(logx.FieldRecover, fmt.Sprintf("%v", r)).
					Error()

				resp.Error(ctx, errorx.ErrInternal)
			}
		}()

		ctx.Next()
	}
}

func topFrame() (fn, file string, line int) {
	const maxDepth = 64
	var pcs [maxDepth]uintptr
	// 可以跳过前三帧，前三帧依次是 Callers 本身、topFrame、Recovery 的 defer
	n := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	for {
		f, more := frames.Next()
		if !strings.Contains(f.File, "/runtime") &&
			!strings.Contains(f.File, "/recovery") {
			return f.Function, f.File, f.Line
		}
		if !more {
			return "", "", 0
		}
	}
}

func filterHeaders(h http.Header, hideHeaders []string) map[string]string {
	hideHeaderSet := make(map[string]struct{})
	for _, header := range hideHeaders {
		hideHeaderSet[header] = struct{}{}
	}

	filtered := make(map[string]string, len(h))
	for k, v := range h {
		lk := strings.ToLower(k)
		if _, deny := hideHeaderSet[lk]; deny {
			filtered[k] = "***"
		} else {
			filtered[k] = strings.Join(v, ",")
		}
	}
	return filtered
}

// 只适用于非 windows 平台 go:build !windows
func isBrokenPipe(recovered any, request *http.Request) bool {
	if request != nil && request.Context().Err() == context.Canceled {
		return true
	}

	err, ok := recovered.(error)
	if !ok || err == nil {
		return false
	}

	if errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) {
		return true
	}

	var ne *net.OpError
	if errors.As(err, &ne) {
		var se *os.SyscallError
		if errors.As(ne.Err, &se) {
			if errors.Is(se.Err, syscall.EPIPE) || errors.Is(se.Err, syscall.ECONNRESET) {
				return true
			}
		}
	}

	s := strings.ToLower(err.Error())
	if strings.Contains(s, "broken pipe") ||
		strings.Contains(s, "connection reset by peer") ||
		strings.Contains(s, "write: client disconnected") ||
		strings.Contains(s, "use of closed network connection") {
		return true
	}

	return false
}

type errCapturingWriter struct {
	gin.ResponseWriter
	lastErr atomic.Value
}

func (w *errCapturingWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	if err != nil {
		w.lastErr.Store(err)
	}
	return n, err
}

func (w *errCapturingWriter) WriteString(s string) (int, error) {
	n, err := w.ResponseWriter.WriteString(s)
	if err != nil {
		w.lastErr.Store(err)
	}
	return n, err
}

func (w *errCapturingWriter) LastErr() error {
	v := w.lastErr.Load()
	if v == nil {
		return nil
	}
	return v.(error)
}
