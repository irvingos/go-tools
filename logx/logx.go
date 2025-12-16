package logx

import (
	"context"

	"github.com/irvingos/go-tools/trace"
	"github.com/sirupsen/logrus"
)

var rootEntry *logrus.Entry

func Init(o *Options) {
	o.normalize()

	base := logrus.New()
	switch o.Format {
	case FormatText:
		base.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: string(o.TimestampFormat),
			FullTimestamp:   true,
		})
	case FormatJson:
		base.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: string(o.TimestampFormat),
		})
	}
	base.SetLevel(o.Level)
	base.SetOutput(o.Output)
	for _, hook := range o.Hooks {
		base.AddHook(hook)
	}

	rootEntry = logrus.NewEntry(base)
}

type E struct {
	*logrus.Entry
}

func (e *E) WithField(key Field, val any) *E {
	e.Entry = e.Entry.WithField(key, val)
	return e
}

func (e *E) WithFields(fields map[Field]any) *E {
	e.Entry = e.Entry.WithFields(fields)
	return e
}

func (e *E) WithCaller(caller string) *E {
	e.Entry = e.Entry.WithField(FieldCaller, caller)
	return e
}

func (e *E) WithError(err error) *E {
	e.Entry = e.Entry.WithError(err)
	return e
}

func (e *E) withTrace(ctx context.Context) *E {
	traceID := trace.TraceIDFrom(ctx)
	if traceID == "" {
		return e
	}
	e.Entry = e.Entry.WithField(FieldTraceID, traceID)
	return e
}

func WithContext(ctx context.Context) *E {
	e := E{rootEntry.WithContext(ctx)}
	return e.withTrace(ctx).WithCaller(defaultCaller())
}

func Info(args ...any) {
	WithContext(context.Background()).WithCaller(defaultCaller()).Info(args...)
}

func Infof(format string, args ...any) {
	WithContext(context.Background()).WithCaller(defaultCaller()).Infof(format, args...)
}

func Error(args ...any) {
	WithContext(context.Background()).WithCaller(defaultCaller()).Error(args...)
}

func Errorf(format string, args ...any) {
	WithContext(context.Background()).WithCaller(defaultCaller()).Errorf(format, args...)
}

func Warn(args ...any) {
	WithContext(context.Background()).WithCaller(defaultCaller()).Warn(args...)
}

func Warnf(format string, args ...any) {
	WithContext(context.Background()).WithCaller(defaultCaller()).Warnf(format, args...)
}

func Fatal(args ...any) {
	WithContext(context.Background()).WithCaller(defaultCaller()).Fatal(args...)
}

func Fatalf(format string, args ...any) {
	WithContext(context.Background()).WithCaller(defaultCaller()).Fatalf(format, args...)
}
