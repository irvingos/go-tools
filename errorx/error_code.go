package errorx

import "fmt"

var (
	ErrBadRequest          = NewError(400, "common.bad_request")
	ErrUnauthorized        = NewError(401, "common.unauthorized")
	ErrForbidden           = NewError(403, "common.forbidden")
	ErrNotFound            = NewError(404, "common.not_found")
	ErrInternalServerError = NewError(500, "common.internal_server_error")
)

var errorCodes = make(map[int]Error)

type errorCode struct {
	code    int
	i18nKey string
}

func (e errorCode) Error() string {
	return fmt.Sprintf("code: %d, i18nKey: %s", e.code, e.i18nKey)
}

// Code implements [Code].
func (e errorCode) Code() int {
	return e.code
}

// I18nKey implements [Error].
func (e errorCode) I18nKey() string {
	return e.i18nKey
}

func NewError(code int, i18nKey string) Error {
	if _, ok := errorCodes[code]; ok {
		panic(fmt.Sprintf("error code %d already registered", code))
	}
	errorCodes[code] = &errorCode{code: code, i18nKey: i18nKey}
	return errorCodes[code]
}
