package logx

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

type traceSQLKey struct{}

func WithTraceSQL(ctx context.Context) context.Context {
	return context.WithValue(ctx, traceSQLKey{}, true)
}

func isTraceSQL(ctx context.Context) bool {
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

// 默认的方法直接调用 Entry(ctx, utils.FileWithLineNum()).XXX，由 logx 完成 TraceID 输出
func (l *tracedDBLogger) Error(ctx context.Context, msg string, data ...any) {
	WithContext(ctx).WithCaller(utils.FileWithLineNum()).Errorf(msg, data...)
}

func (l *tracedDBLogger) Info(ctx context.Context, msg string, data ...any) {
	WithContext(ctx).WithCaller(utils.FileWithLineNum()).Infof(msg, data...)
}

func (l *tracedDBLogger) Warn(ctx context.Context, msg string, data ...any) {
	WithContext(ctx).WithCaller(utils.FileWithLineNum()).Warnf(msg, data...)
}

// Trace 方法对输出进行定制，输出 gorm 提供的 SQL 调用方
func (l *tracedDBLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.level == logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()
	// caller 必须在这里获取，不能是在 emit 方法里获取，否则 caller 将会是 db_logger.go
	caller := utils.FileWithLineNum()

	isErr := err != nil && !errors.Is(err, gorm.ErrRecordNotFound)
	isSlow := l.SlowSQLThreshold != 0 && elapsed > l.SlowSQLThreshold

	if isTraceSQL(ctx) {
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
