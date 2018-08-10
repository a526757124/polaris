package balance

import (
	"github.com/devfeel/polaris/util/logx"
	"github.com/devfeel/polaris/models"
	"github.com/devfeel/polaris/const"
	"math/rand"
	"strconv"
	"time"
)

var (
	LBLogger logger.Logger
)

func init() {
	LBLogger = logger.LoadBalanceLogger
}

/* 获取目标Api，对外屏蔽负载算法
* Author: Panxinming
* LastUpdateTime: 2016-08-30 18:00
* 目前使用随机负载
* 如果合法，返回具体的某一个ApiUrl
 */
func GetTargetApi(apiInfo *models.GatewayApiInfo) *models.TargetApiInfo {
	if apiInfo.ApiType == _const.ApiType_Group {
		return nil
	}
	api := getRandTarget(apiInfo.TargetApi)
	if api != nil{
		LBLogger.Info("[" + strconv.Itoa(apiInfo.ApiID) + "] GetTargetApi=>" + api.TargetUrl)
	}else{
		LBLogger.Info("[" + strconv.Itoa(apiInfo.ApiID) + "] GetTargetApi nil")
	}

	return api
}

//随机从字符串数组里获取一项
func getRandTarget(targetApis []*models.TargetApiInfo) *models.TargetApiInfo {
	valLen := len(targetApis)
	if valLen <= 0 {
		return nil
	} else if valLen == 1 {
		return targetApis[0]
	} else {
		rand.Seed(time.Now().UnixNano())
		randIndex := rand.Intn(valLen)
		return targetApis[randIndex]
	}
}
