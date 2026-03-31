package errorx

import (
	"context"
	"errors"
)

func IsCanceled(err error) bool {
	return errors.Is(err, context.Canceled)
}

func IsTimeout(err error) bool {
	return errors.Is(err, context.DeadlineExceeded)
}

func IsContextDone(err error) bool {
	return IsCanceled(err) || IsTimeout(err)
}
