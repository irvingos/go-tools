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

type tracedDBLogger struct {
	slowSQLThreshold time.Duration
}

func NewDBLogger(slowSQLThreshold time.Duration) *tracedDBLogger {
	return &tracedDBLogger{
		slowSQLThreshold: slowSQLThreshold,
	}
}

func (l *tracedDBLogger) LogMode(level logger.LogLevel) logger.Interface {
	// do nothing
	return l
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
	elapsed := time.Since(begin)
	sql, rowsAffected := fc()

	logEntry := WithContext(ctx).
		WithCaller(utils.FileWithLineNum()).
		WithField("sql", strings.ReplaceAll(sql, "\"", "'")).
		WithField("rows", rowsOrDash(rowsAffected)).
		WithField("elapsed", fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6))

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logEntry = logEntry.WithField("error", err)
	}
	if elapsed > l.slowSQLThreshold && l.slowSQLThreshold != 0 {
		logEntry = logEntry.WithField("slow", true)
	}

	logEntry.Info("gorm")
}

func rowsOrDash(rows int64) any {
	if rows == -1 {
		return "-"
	}
	return rows
}
