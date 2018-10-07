package viewModel

import (
	"github.com/a526757124/polaris/uiserver/common/page"
)

func init() {

}

//应用管理列表查询参数
type APIGroupQueryParm struct {
	//分页对象
	page.PageParam
	//应用名称
	GroupName string
}
