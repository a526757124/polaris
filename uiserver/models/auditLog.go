package models

//审计日志
type AuditLog struct {
	ID             int64
	OperationUser  int64  //操作人
	OperationTime  int64  //操作时间
	BusinessModule int64  //业务模块
	UserAgent      string //客户端信息
	OldData        string //修改前数据
	NewData        string //修改后数据
	CreateTime     int64  //创建时间
}
