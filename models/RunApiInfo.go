package models

import (
	"time"
)

//目标Api信息
type RunApiInfo struct {
	//Api编号
	ApiID int
	//Api对应的真实Url
	ApiUrl string
	//累计失败次数
	TotalErrorCount int
	//最后检查时间
	LastValidateTime time.Time
}
