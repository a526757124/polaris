package models

//应用信息
type AppInfo struct {
	ID          int64
	Name        string //名称
	Desc        string //描述
	Key         string //密钥
	Url         string //地址
	IPList      string //服务器IP
	DevUser     string //开发人员
	ProductUser string //产品人员
	Status      int64  //状态 0:是初始化 1:是有效 -1:是无效
	CreateUser  int64  //创建人
	CreateTime  int64  //创建时间
}
