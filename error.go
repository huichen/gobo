// 自定义错误结构体
package gobo

type ErrorString struct {
	s string
}

func (e *ErrorString) Error() string {
	return e.s
}
