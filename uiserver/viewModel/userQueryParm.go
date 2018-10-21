package viewModel

import "github.com/a526757124/polaris/uiserver/common/page"

func init() {

}

//用户管理列表查询参数
type UserQueryParm struct {
	//分页对象
	page.PageParam
	//用户昵称
	NickName string
}
