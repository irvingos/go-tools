package resp

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/irvingos/go-tools/errorx"
)

var (
	errEmptyParam    = errorx.NewError(1004100, "empty request param")
	errValidateParam = errorx.NewError(1004101, "error validate param")
	errResolveParam  = errorx.NewError(1004102, "error resolve param")
)

type Response struct {
	Code    int    `json:"code"`
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Detail  string `json:"detail,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func OK(g *gin.Context, data any) {
	g.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

func ErrorParam(g *gin.Context, err error) {
	if errors.Is(err, io.EOF) {
		g.AbortWithStatusJSON(http.StatusOK, Response{
			Code:    errEmptyParam.Code(),
			Message: errEmptyParam.Message(),
			Detail:  err.Error(),
		})
		WithCode(g, errEmptyParam.Code())
		return
	}
	var ve validator.ValidationErrors
	if ok := errors.As(err, &ve); ok {
		g.AbortWithStatusJSON(http.StatusOK, Response{
			Code:    errValidateParam.Code(),
			Message: errValidateParam.Message(),
			Detail:  err.Error(),
		})
		WithCode(g, errValidateParam.Code())
		return
	}
	g.AbortWithStatusJSON(http.StatusOK, Response{
		Code:    errResolveParam.Code(),
		Message: errResolveParam.Message(),
		Detail:  err.Error(),
	})
	WithCode(g, errResolveParam.Code())
}

func Error(g *gin.Context, err error) {
	if apiErr, ok := err.(errorx.Error); ok {
		g.AbortWithStatusJSON(http.StatusOK, Response{
			Code:    apiErr.Code(),
			Message: apiErr.Message(),
		})
		WithCode(g, apiErr.Code())
		return
	}

	g.AbortWithStatusJSON(http.StatusOK, Response{
		Code:    errorx.ErrInternal.Code(),
		Message: errorx.ErrInternal.Message(),
		Detail:  err.Error(),
	})
	WithCode(g, errorx.ErrInternal.Code())
}
