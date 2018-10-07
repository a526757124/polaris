package viewModel

func init() {
	//mapper.Register(AppInfoDto{})
	//conn.GetRedisClient().SetNX("appincr", "10000")
}

//应用管理列表
type AppInfoDto struct {
	//App编号
	AppID int64 `mapper:"ID"`
	//应用名称
	AppName string `mapper:"Name"`
	//应用描述
	AppDesc string `mapper:"Desc"`
	//应用地址
	AppUrl string
	//应用服务器IP
	AppIPList string
	//开发人员
	DevUser string
	//产品人员
	ProductUser string
	//应用加密Key
	AppKey string
	//Api状态 0有效
	Status int
}

// TableName 表名
func (*AppInfoDto) TableName() string {
	return "app"
}
