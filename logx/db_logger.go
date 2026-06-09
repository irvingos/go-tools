package logx

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type traceSQLKey struct{}

func WithTraceSQL(ctx context.Context) context.Context {
	return context.WithValue(ctx, traceSQLKey{}, true)
}

func IsTraceSQL(ctx context.Context) bool {
	trace, ok := ctx.Value(traceSQLKey{}).(bool)
	return ok && trace
}

type DBLoggerOptions struct {
	SlowSQLThreshold time.Duration
}
type tracedDBLogger struct {
	DBLoggerOptions

	level logger.LogLevel
}

func NewDBLogger(o *DBLoggerOptions) logger.Interface {
	return &tracedDBLogger{
		DBLoggerOptions: *o,
		level:           logger.Warn,
	}
}

func (l *tracedDBLogger) LogMode(level logger.LogLevel) logger.Interface {
	nL := *l
	nL.level = level
	return &nL
}

// 默认的方法直接调用 Entry(ctx, sqlCaller()).XXX，由 logx 完成 TraceID 输出
func (l *tracedDBLogger) Error(ctx context.Context, msg string, data ...any) {
	WithContext(ctx).WithCaller(sqlCaller()).Errorf(msg, data...)
}

func (l *tracedDBLogger) Info(ctx context.Context, msg string, data ...any) {
	WithContext(ctx).WithCaller(sqlCaller()).Infof(msg, data...)
}

func (l *tracedDBLogger) Warn(ctx context.Context, msg string, data ...any) {
	WithContext(ctx).WithCaller(sqlCaller()).Warnf(msg, data...)
}

// Trace 方法对输出进行定制，输出 gorm 提供的 SQL 调用方
func (l *tracedDBLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.level == logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()
	// caller 必须在这里获取，不能是在 emit 方法里获取，否则 caller 将会是 db_logger.go
	caller := sqlCaller()

	isErr := err != nil && !errors.Is(err, gorm.ErrRecordNotFound) && !errors.Is(err, context.Canceled)
	isSlow := l.SlowSQLThreshold != 0 && elapsed > l.SlowSQLThreshold

	if IsTraceSQL(ctx) {
		l.emit(ctx, sql, caller, rows, elapsed, isSlow, err)
		return
	}

	shouldLog := false
	switch l.level {
	case logger.Info:
		shouldLog = true
	case logger.Warn:
		shouldLog = isErr || isSlow
	case logger.Error:
		shouldLog = isErr
	default:
		shouldLog = isErr || isSlow
	}

	if !shouldLog {
		return
	}

	l.emit(ctx, sql, caller, rows, elapsed, isSlow, err)
}

func (l *tracedDBLogger) emit(ctx context.Context, sql, caller string, rows int64, elapsed time.Duration, isSlow bool, err error) {
	entry := WithContext(ctx).
		WithCaller(caller).
		WithField("sql", strings.ReplaceAll(sql, "\"", "'")).
		WithField("rows", rowsOrDash(rows)).
		WithField("elapsed", fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6))

	if isSlow {
		entry = entry.WithField("slow", true)
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		entry = entry.WithField("error", err)
	}

	entry.Info("gorm")
}

func rowsOrDash(rows int64) any {
	if rows == -1 {
		return "-"
	}
	return rows
}

// sqlCaller 复用 gorm logger 的栈回溯策略，但额外跳过 logx 包装层。
// gorm 内置 logger 在 Trace 内直接调用 utils.FileWithLineNum()（skip=3）；
// tracedDBLogger 多一层 Trace 包装，若仍用 utils.FileWithLineNum() 会把 caller 解析到 db_logger.go。
func sqlCaller() string {
	pcs := [13]uintptr{}
	n := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		if frame.PC == 0 {
			break
		}
		if isSQLCallerFrame(frame) {
			return string(strconv.AppendInt(append([]byte(frame.File), ':'), int64(frame.Line), 10))
		}
		if !more {
			break
		}
	}
	return ""
}

func isSQLCallerFrame(frame runtime.Frame) bool {
	file := frame.File
	if strings.HasSuffix(file, "_test.go") || strings.HasSuffix(file, ".gen.go") {
		return false
	}
	if strings.Contains(file, "gorm.io/gorm") || strings.Contains(file, "gorm.io/gen") {
		return false
	}
	if strings.Contains(file, "/logx/") {
		return false
	}
	return true
}
