package errorx

import (
	"encoding/json"
	"fmt"
)

type Error interface {
	error
	Code() int
	Message() string
	HTTPStatus() int
}

type errorx struct {
	code    int
	message string
	status  int
}

func (e errorx) Error() string {
	return e.message
}

func (e errorx) Code() int {
	return e.code
}

func (e errorx) Message() string {
	return e.message
}

func (e errorx) HTTPStatus() int {
	return e.status
}

func (e errorx) MarshalJSON() ([]byte, error) {
	type alias struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	return json.Marshal(alias{
		Code:    e.code,
		Message: e.message,
	})
}

func NewError(code int, message string, status int) Error {
	return errorx{code: code, message: message, status: status}
}

func Errorf(err Error, args ...any) Error {
	return errorx{code: err.Code(), message: fmt.Sprintf(err.Message(), args...), status: err.HTTPStatus()}
}
