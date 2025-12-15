package logx

import (
	"io"
	"os"

	"github.com/irvingos/go-tools/consts"

	"github.com/sirupsen/logrus"
)

type Format string

const (
	FormatText Format = "text"
	FormatJson Format = "json"
)

type Level = logrus.Level

type Options struct {
	Format          Format
	TimestampFormat consts.TimeFormat
	Level           logrus.Level
	Output          io.Writer
	Hooks           []logrus.Hook
}

func (o *Options) normalize() {
	// fill defaults
	if o.Format == "" {
		o.Format = FormatText
	}
	if o.TimestampFormat == "" {
		o.TimestampFormat = consts.TimeToSecondFormat
	}
	if o.Level == 0 {
		o.Level = logrus.InfoLevel
	}
	if o.Output == nil {
		o.Output = os.Stdout
	}
}
