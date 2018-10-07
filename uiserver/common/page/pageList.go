package page

//分页查询后返回数据结构
type PageList struct {
	//总条数
	TotalCount int64 `json:"totalCount"`
	//当前页数据集合
	PageData interface{} `json:"pageData"`
}
