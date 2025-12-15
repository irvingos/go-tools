package errorx

import (
	"net/http"
)

var (
	ErrBadRequest   = NewError(1004000, "bad request", http.StatusBadRequest)
	ErrUnauthorized = NewError(1004001, "unauthorized", http.StatusUnauthorized)
	ErrInvalidToken = NewError(1004002, "invalid token", http.StatusBadRequest)
	ErrForbidden    = NewError(1004003, "forbidden", http.StatusForbidden)
	ErrNotFound     = NewError(1004004, "resource not found", http.StatusNotFound)

	ErrInternal = NewError(1005000, "internal server error", http.StatusInternalServerError)
)
