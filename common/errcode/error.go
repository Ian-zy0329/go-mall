package errcode

import (
	"encoding/json"
	"fmt"
	"path"
	"runtime"
)

type AppError struct {
	code     int    `json:"code"`
	msg      string `json:"msg"`
	cause    error  `json:"cause"`
	occurred string `json:"occurred"`
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	formattedErr := struct {
		Code  int    `json:"code"`
		Msg   string `json:"msg"`
		Cause string `json:"cause"`
	}{
		Code: e.code,
		Msg:  e.msg,
	}
	if e.cause != nil {
		formattedErr.Cause = e.cause.Error()
	}
	errByte, _ := json.Marshal(formattedErr)
	return string(errByte)
}

func (e *AppError) String() string {
	return e.Error()
}

func newError(code int, msg string) *AppError {
	if code > -1 {
		if _, duplicated := codes[code]; duplicated {
			panic(fmt.Sprintf("预定义错误码 %d 不能重复, 请检查后更换", code))
		}
		codes[code] = struct{}{}
	}

	return &AppError{
		code: code,
		msg:  msg,
	}
}

func (e *AppError) Code() int {
	return e.code
}

func (e *AppError) Msg() string {
	return e.msg
}

func Wrap(msg string, err error) *AppError {
	if err == nil {
		return nil
	}
	appErr := &AppError{code: -1, msg: msg, cause: err}
	appErr.occurred = getAppErrOccurredInfo()
	return appErr
}

func (e *AppError) WithCause(err error) *AppError {
	newErr := e.Clone()
	newErr.cause = err
	newErr.occurred = getAppErrOccurredInfo()
	return newErr
}

func (e *AppError) Clone() *AppError {
	return &AppError{
		code:     e.code,
		msg:      e.msg,
		cause:    e.cause,
		occurred: e.occurred,
	}
}

func getAppErrOccurredInfo() string {
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		return ""
	}
	file = path.Base(file)
	funcName := runtime.FuncForPC(pc).Name()
	triggerInfo := fmt.Sprintf("func: %s, file: %s, line: %d", funcName, file, line)
	return triggerInfo
}

func (e *AppError) Unwrap() error {
	return e.cause
}

func (e *AppError) Is(target error) bool {
	targetErr, ok := target.(*AppError)
	if !ok {
		return false
	}
	return targetErr.Code() == e.Code()
}
