package models

//应用与api关联
type AppRelationAPI struct {
	ID         int64
	AppID      int64
	ApiID      int64
	IsUse      bool  //是否使用
	CreateUser int64 //创建人
	CreateTime int64
}
