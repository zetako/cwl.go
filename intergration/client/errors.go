package client

import "fmt"

const (
	CodeFileNotExist int = 11502
)

type StarlightResponseError struct {
	code int
	msg  string
}

func NewError(code int, msg string) error {
	return StarlightResponseError{
		code: code,
		msg:  msg,
	}
}

func (e StarlightResponseError) Error() string {
	return fmt.Sprintf("(%d)%s", e.code, e.msg)
}

func MustCode(err error) int {
	switch serr := err.(type) {
	case StarlightResponseError:
		return serr.code
	default:
		return -1
	}
}
