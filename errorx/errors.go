package errorx

var (
	ErrBadRequest   = NewError(1004000, "bad request")
	ErrUnauthorized = NewError(1004001, "unauthorized")
	ErrInvalidToken = NewError(1004002, "invalid token")
	ErrForbidden    = NewError(1004003, "forbidden")
	ErrNotFound     = NewError(1004004, "resource not found")

	ErrInternal = NewError(1005000, "internal server error")
)
