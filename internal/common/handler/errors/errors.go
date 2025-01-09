package errors

import (
	"errors"
	"fmt"
	"github.com/baobao233/gorder/common/consts"
)

type Error struct {
	code int
	msg  string
	err  error
}

func (e *Error) Error() string {
	var msg string
	if e.msg != "" {
		msg = e.msg
	}
	msg = consts.ErrMsg[e.code]
	return msg + " -> " + e.err.Error()
}

func New(code int) error {
	return &Error{
		code: code,
	}
}

func NewWithError(code int, err error) error {
	if err == nil {
		return New(code)
	}
	return &Error{
		code: code,
		err:  err,
	}
}

func NewWithMsg(code int, format string, args ...any) error {
	return &Error{
		code: code,
		msg:  fmt.Sprintf(format, args...),
	}
}

func Errno(err error) int {
	// 判断是否有 error，没有则返回成功
	if err == nil {
		return consts.ErrnoSuccess
	}
	// 判断是不是我们自己定义的 error
	targetError := &Error{}
	if ok := errors.As(err, &targetError); ok {
		return targetError.code
	}
	return -1
}

func Output(err error) (int, string) {
	if err == nil {
		return consts.ErrnoSuccess, consts.ErrMsg[consts.ErrnoSuccess]
	}
	errno := Errno(err)
	if errno == -1 {
		return consts.ErrnoUnknown, err.Error()
	}
	return errno, err.Error()
}
