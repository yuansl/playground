package httputil

import (
	"fmt"
)

var (
	ErrInvalid = &Error{Code: 400001, Message: "Invalid argument"}
)

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("http error: code: %d, message: %s, data: %q", e.Code, e.Message, e.Data)
}

func (e *Error) Clone() Error {
	return *e
}

func NewError(code int, message string, opts ...any) *Error {
	return &Error{Code: code, Message: message, Data: opts}
}

func NewErrorWith(e *Error, optional ...any) *Error {
	var err = e.Clone()
	var detail []any

	for _, op := range optional {
		if _e, ok := op.(error); ok {
			detail = append(detail, _e.Error())
		} else if desc, ok := op.(fmt.Stringer); ok {
			detail = append(detail, desc.String())
		}
	}
	err.Data = detail

	return &err
}
