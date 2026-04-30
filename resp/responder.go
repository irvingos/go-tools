package resp

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/irvingos/go-tools/errorx"
	"github.com/irvingos/go-tools/i18n"
)

func NewResponder(catalog *i18n.Catalog) *Responder {
	return &Responder{catalog: catalog}
}

type Response struct {
	Code    int    `json:"code"`
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Detail  string `json:"detail,omitempty"`
	Data    any    `json:"data,omitempty"`
}

type Responder struct {
	catalog *i18n.Catalog
}

func (r *Responder) Success(g *gin.Context, data any) {
	g.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

func (r *Responder) Error(g *gin.Context, err error) {
	ctx := g.Request.Context()

	var errorCode errorx.Error
	if ok := errors.As(err, &errorCode); ok {
		msg := r.catalog.ErrorMessage(ctx, err)

		g.AbortWithStatusJSON(http.StatusOK, Response{
			Code:    errorCode.Code(),
			Message: msg,
		})
		g.Request = g.Request.WithContext(WithCode(ctx, errorCode.Code()))
		return
	}

	msg := r.catalog.ErrorMessage(ctx, errorx.ErrInternalServerError)
	g.AbortWithStatusJSON(http.StatusOK, Response{
		Code:    errorx.ErrInternalServerError.Code(),
		Message: msg,
		Detail:  err.Error(),
	})
	g.Request = g.Request.WithContext(WithCode(ctx, errorx.ErrInternalServerError.Code()))
}

func (r *Responder) ErrorParam(g *gin.Context, err error) {
	ctx := g.Request.Context()

	msg := r.catalog.ErrorMessage(ctx, errorx.ErrBadRequest)
	g.AbortWithStatusJSON(http.StatusOK, Response{
		Code:    errorx.ErrBadRequest.Code(),
		Message: msg,
		Detail:  err.Error(),
	})
	g.Request = g.Request.WithContext(WithCode(ctx, errorx.ErrBadRequest.Code()))
}
