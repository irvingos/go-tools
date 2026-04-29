package errorx

import (
	"errors"
	"testing"
)

func TestFormat(t *testing.T) {
	err := Format(ErrBadRequest, "name")

	if got, want := err.Error(), ErrBadRequest.Error(); got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
	if got, want := err.Code(), ErrBadRequest.Code(); got != want {
		t.Fatalf("Code() = %d, want %d", got, want)
	}
	if got, want := err.I18nKey(), ErrBadRequest.I18nKey(); got != want {
		t.Fatalf("I18nKey() = %q, want %q", got, want)
	}
	if !errors.Is(err, ErrBadRequest) {
		t.Fatalf("errors.Is(err, ErrBadRequest) = false, want true")
	}

	var errorCode Error
	if !errors.As(err, &errorCode) {
		t.Fatalf("errors.As(err, &errorCode) = false, want true")
	}
	if errorCode.Code() != ErrBadRequest.Code() {
		t.Fatalf("errors.As code = %d, want %d", errorCode.Code(), ErrBadRequest.Code())
	}

	var argError interface {
		Args() []any
	}
	if !errors.As(err, &argError) {
		t.Fatalf("errors.As(err, &argError) = false, want true")
	}
	args := argError.Args()
	if len(args) != 1 || args[0] != "name" {
		t.Fatalf("Args() = %#v, want []any{\"name\"}", args)
	}
}
