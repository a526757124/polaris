// AppInfo
package models

//应用信息
type AppInfo struct {
	//App编号
	AppID int
	//应用名称
	AppName string
	//应用加密Key
	AppKey string
	//Api状态 0有效
	Status int
}
