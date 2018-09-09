package page

//分页参数
type PageParam struct {
	//当前页码
	PageIndex int
	//每页显示总条数
	PageSize int
	//排序字段
	SortList []*SortDto
}

//排序
type SortDto struct {
	//排序字段
	SortField string
	//排序方式 asc\desc
	SortWay string
}
