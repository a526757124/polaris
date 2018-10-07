package page

//分页参数
type PageParam struct {
	//当前页码
	PageIndex int64 `json:"pageIndex"`
	//每页显示总条数
	PageSize int64 `json:"pageSize"`
	//排序字段
	SortList []*SortDto `json:"sortList"`
}

//排序
type SortDto struct {
	//排序字段
	SortField string `json:"sortField"`
	//排序方式 asc\desc
	SortWay string `json:"sortWay"`
}

//get skip
func (page *PageParam) GetSkip() int64 {
	if page.PageIndex <= 0 {
		page.PageIndex = 1
	}
	return (page.PageIndex - 1) * page.PageSize
}

//get limit
func (page *PageParam) GetLimit() int64 {
	return page.PageSize
}
