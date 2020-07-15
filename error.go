package easycall

import "strconv"

const (
	ERROR_METHOD_NOT_FOUND  = 1002
	ERROR_SERVICE_NOT_FOUND = 1002
	ERROR_INTERNAL_ERROR    = 1001
	ERROR_TIME_OUT          = 1003
)

type LogicError struct {
	ret int
	msg string
}

func NewLogicError(ret int, msg string) *LogicError {
	return &LogicError{ret, msg}
}

func (e *LogicError) Error() string {
	return "Ret=" + strconv.Itoa(e.ret) + ",Msg=" + e.msg
}

func (e *LogicError) GetRet() int {
	return e.ret
}

func (e *LogicError) GetMsg() string {
	return e.msg
}

type SystemError struct {
	ret int
	msg string
}

func NewSystemError(ret int, msg string) *SystemError {
	return &SystemError{ret, msg}
}

func (e *SystemError) Error() string {
	return "Ret=" + strconv.Itoa(e.ret) + ",Msg=" + e.msg
}

func (e *SystemError) GetRet() int {
	return e.ret
}

func (e *SystemError) GetMsg() string {
	return e.msg
}
