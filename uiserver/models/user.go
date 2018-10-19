package models

//登录用户
type User struct {
	ID         int64  //ID
	NickName   string //用户昵称
	LoginName  string //登录名
	LoginPwd   string //登录密码
	Status     int64  //状态 0:是初始化 1:是有效 -1:是无效
	CreateTime int64  //创建时间
}
