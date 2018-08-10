// AppApiRelation
package models

//应用与API关系实体，表示指定APP是否拥有指定API的调用权限
type Relation struct {
	//App编号
	AppID int
	//应用名称
	ApiID int
	//应用加密Key
	IsUse bool
}
