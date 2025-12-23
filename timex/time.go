package timex

import "time"

type Format string

const (
	Second Format = "2006-01-02 15:04:05"
)

func (f Format) Format(t time.Time) string {
	return t.Format(string(f))
}

func (f Format) String() string {
	return string(f)
}
