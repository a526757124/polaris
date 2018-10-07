package viewModel

import (
	"github.com/a526757124/polaris/uiserver/conn"
)

func init() {
	//mapper.Register(AppInfoDto{})
	conn.GetRedisClient().SetNX("apiGroupincr", "10000")
}

//分组管理列表
type APIGroupDto struct {
	GroupID   int64
	GroupName string
	GroupDesc string
}

// TableName 表名
func (*APIGroupDto) TableName() string {
	return "apiGroup"
}
