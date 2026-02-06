package reply

import (
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// grpc跟业务异常的映射关系
var GrpcBusinessErrorCodes = map[codes.Code]int32{}

// 业务异常跟grpc的映射关系
var BusinessGrpcErrorCodes = map[int32]codes.Code{}

type Error struct {
	BaseResponse
	Details []interface{}
	Cause   *Error
}

func (e *Error) Error() string {
	var result string
	if len(e.Details) > 0 {
		result = fmt.Sprintf("error code: %d, message: %s, details: %v", e.Code, e.Message, e.Details)
	} else {
		result = fmt.Sprintf("error code: %d, message: %s", e.Code, e.Message)
	}

	if e.Cause != nil {
		result += fmt.Sprintf(", cause: %v", e.Cause)
	}

	return result
}

func New(code int32, message string, details ...any) *Error {
	return &Error{
		BaseResponse: BaseResponse{
			Code:    code,
			Message: message,
		},
		Details: details,
	}
}

func Newf(code int32, format string, args ...any) *Error {
	return &Error{
		BaseResponse: BaseResponse{
			Code:    code,
			Message: fmt.Sprintf(format, args...),
		},
	}
}

func (e *Error) WithDetails(details ...any) {
	e.Details = append(e.Details, details...)
}

func (e *Error) WithCause(cause *Error) {
	e.Cause = cause
}

func (e *Error) Unwrap() *Error {
	return e.Cause
}

// Wrap 包装现有错误，创建新的 Error
func Wrap(err *Error, code int32, message string, details ...any) *Error {
	return &Error{
		BaseResponse: BaseResponse{
			Code:    code,
			Message: message,
		},
		Details: details,
		Cause:   err,
	}
}

func Wrapf(err *Error, code int32, format string, args ...any) *Error {
	return &Error{
		BaseResponse: BaseResponse{
			Code:    code,
			Message: fmt.Sprintf(format, args...),
		},
		Cause: err,
	}
}

func RelationshipMaintenance(code codes.Code, errCode int32) {
	GrpcBusinessErrorCodes[code] = errCode
	BusinessGrpcErrorCodes[errCode] = code
}

// ToGRPCError 转换为 rpc 错误
func (e *Error) ToGRPCError() error {
	errorCodes := BusinessGrpcErrorCodes[e.Code]
	return status.Error(errorCodes, e.Message)
}

// FromGRPCError 从 rpc 错误转换
func FromGRPCError(err error) *Error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return New(-1, err.Error())
	}

	errCode := GrpcBusinessErrorCodes[st.Code()]
	return New(errCode, st.Message())
}

func (e *Error) Is(target error) bool {
	var t *Error
	if errors.As(target, &t) {
		return e.Code == t.Code
	}
	return false
}

func WrapError(err *Error, message string, details ...any) *Error {
	if err == nil {
		return nil
	}

	newErr := &Error{
		BaseResponse: BaseResponse{
			Code:    err.Code,
			Message: message,
		},
		Details: details,
		Cause:   err,
	}
	return newErr
}

func WrapErrorf(err *Error, format string, args ...any) *Error {
	if err == nil {
		return nil
	}

	newErr := &Error{
		BaseResponse: BaseResponse{
			Code:    err.Code,
			Message: fmt.Sprintf(format, args...),
		},
		Cause: err,
	}
	return newErr
}

func WrapErrorWithDetails(err *Error, details ...any) *Error {
	if err == nil {
		return nil
	}

	newErr := &Error{
		BaseResponse: BaseResponse{
			Code:    err.Code,
			Message: err.Message,
		},
		Details: details,
		Cause:   err,
	}
	return newErr
}

func WrapErrorWithContext(err *Error, context string, details ...any) *Error {
	if err == nil {
		return nil
	}

	newErr := &Error{
		BaseResponse: BaseResponse{
			Code:    err.Code,
			Message: fmt.Sprintf("%s: %s", context, err.Message),
		},
		Details: details,
		Cause:   err,
	}
	return newErr
}
