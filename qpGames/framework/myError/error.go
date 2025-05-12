package myError

import (
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Error struct {
	Code int
	Err  error
}

func (e *Error) Error() string {
	return e.Err.Error()
}

func NewError(code int, err error) *Error {
	return &Error{
		Code: code,
		Err:  err,
	}
}

// GrpcError 将自定义错误转换为gRPC错误
func GrpcError(err *Error) error {
	return status.Error(codes.Code(err.Code), err.Err.Error())
}

// ToError 将gRPC错误转换为自定义错误
func ToError(err error) *Error {
	fromError, _ := status.FromError(err)
	return NewError(int(fromError.Code()), errors.New(fromError.Message()))
}
