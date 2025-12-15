package logx

import (
	"fmt"
	"runtime"
)

func defaultCaller() string {
	_, file, line, ok := runtime.Caller(2)
	if ok {
		return fmt.Sprintf("%s:%d", file, line)
	}
	return ""
}
