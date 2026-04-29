package errorx

type FormattedError interface {
	Error
	Args() []any
}

type formattedError struct {
	err  Error
	args []any
}

func (e *formattedError) Error() string {
	return e.err.Error()
}

func (e *formattedError) Code() int {
	return e.err.Code()
}

func (e *formattedError) I18nKey() string {
	return e.err.I18nKey()
}

func (e *formattedError) Unwrap() error {
	return e.err
}

func (e *formattedError) Args() []any {
	return append([]any(nil), e.args...)
}

func Format(err Error, args ...any) Error {
	return &formattedError{
		err:  err,
		args: append([]any(nil), args...),
	}
}
