package balance

import (
	"github.com/devfeel/polaris/util/logx"
	"github.com/devfeel/polaris/models"
	"github.com/devfeel/polaris/const"
	"math/rand"
	"strconv"
	"time"
	"github.com/devfeel/polaris/control/hystrix"
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
func GetAliveApi(apiInfo *models.GatewayApiInfo) *models.TargetApiInfo {
	if apiInfo.ApiType == _const.ApiType_Group {
		return nil
	}
	api := getRandTarget(apiInfo.AliveTargetApis)
	if api != nil{
		LBLogger.Info("[" + strconv.Itoa(apiInfo.ApiID) + "] GetTargetApi=>" + api.TargetKey + ":" + api.TargetUrl + ":" + api.CallName)
	}else{
		LBLogger.Info("[" + strconv.Itoa(apiInfo.ApiID) + "] GetTargetApi nil")
	}
	return api
}

func SetError(apiInfo *models.GatewayApiInfo, apiUrl string){
	apiHystrix, isInit := hystrix.GetHystrix(strconv.Itoa(apiInfo.ApiID) + "^$^" + apiUrl)
	if isInit{
		apiHystrix.SetID(apiUrl)
		apiHystrix.SetExtendedData(apiInfo)
		apiHystrix.RegisterOnTriggerAlive(onTriggerAlive)
		apiHystrix.RegisterOnTriggerHystrix(onTriggerHystrix)
		apiHystrix.Do()
	}
	apiHystrix.GetCounter().Inc(1)
}

//随机从字符串数组里获取一项
func getRandTarget(apiUrls []*models.TargetApiInfo) *models.TargetApiInfo {
	valLen := len(apiUrls)
	if valLen <= 0 {
		return nil
	} else if valLen == 1 {
		return apiUrls[0]
	} else {
		rand.Seed(time.Now().UnixNano())
		randIndex := rand.Intn(valLen)
		return apiUrls[randIndex]
	}
}

func onTriggerAlive(h hystrix.Hystrix){
	apiInfo, isOk := h.GetExtendedData().(*models.GatewayApiInfo)
	if !isOk{
		return
	}
	if apiInfo == nil{
		return
	}
	apiKey := h.GetID()
	var targetApi *models.TargetApiInfo
	isExists := false
	for _, s := range apiInfo.AliveTargetApis {
		if s.TargetKey == apiKey {
			isExists = true
			targetApi = s
			break
		}
	}
	if !isExists{
		//nothing to do
		return
	}
	//add target api into alive apis
	isExists = false
	for _, s := range apiInfo.AliveTargetApis {
		if s == targetApi {
			isExists = true
			break
		}
	}
	if !isExists{
		apiInfo.AliveTargetApis = append(apiInfo.AliveTargetApis, targetApi)
	}
}

func onTriggerHystrix(h hystrix.Hystrix){
	apiInfo, isOk := h.GetExtendedData().(*models.GatewayApiInfo)
	if !isOk{
		return
	}
	if apiInfo == nil{
		return
	}
	apiKey := h.GetID()
	if index := findTargetApiIndex(apiInfo.AliveTargetApis, apiKey);index != -1{
		apiInfo.AliveTargetApis = append(apiInfo.AliveTargetApis[0:index], apiInfo.AliveTargetApis[index+1:len(apiInfo.AliveTargetApis)]...)
	}
}



// FindIndex find slice index if a string is in a set
// if not exists, return -1
func findTargetApiIndex(set []*models.TargetApiInfo, find string) int{
	for index, s := range set {
		if s.TargetKey == find {
			return index
		}
	}
	return -1
}
