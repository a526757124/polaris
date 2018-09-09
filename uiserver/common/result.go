package common

import "github.com/a526757124/polaris/uiserver/common/resultCode"

//前端界面通用返回类型
type Result struct {
	//是否成功
	Success bool
	//返回代码
	Code int
	//返回信息
	Msg string
	//返回数据
	Data interface{}
}

//创建一个返回类型
func NewResult(success bool, code int, msg string, data interface{}) *Result {
	return &Result{
		Success: success,
		Code:    code,
		Msg:     msg,
		Data:    data,
	}
}

//创建一个成功的返回类型
func NewSuccessResult(data interface{}) *Result {
	return &Result{
		Success: true,
		Code:    resultCode.SUCCESSCode,
		Msg:     "成功",
		Data:    data,
	}
}

//创建一个失败的返回类型
func NewFailResult(msg string) *Result {
	return &Result{
		Success: false,
		Code:    resultCode.SUCCESSFail,
		Msg:     msg,
	}
}

//创建一个失败的自定义错误代码返回类型
func NewCustomFailResult(code int, msg string) *Result {
	return &Result{
		Success: false,
		Code:    code,
		Msg:     msg,
	}
}
