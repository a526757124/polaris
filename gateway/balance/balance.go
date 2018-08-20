package balance

import (
	"github.com/devfeel/polaris/util/logx"
	"github.com/devfeel/polaris/models"
	"github.com/devfeel/polaris/const"
	"math/rand"
	"strconv"
	"time"
	"github.com/devfeel/polaris/control/hystrix"
	"github.com/devfeel/polaris/util/slicex"
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
func GetAliveApi(apiInfo *models.GatewayApiInfo) string {
	if apiInfo.ApiType == _const.ApiType_Group {
		return ""
	}
	api := getRandTarget(apiInfo.AliveApiUrls)
	if api != ""{
		LBLogger.Info("[" + strconv.Itoa(apiInfo.ApiID) + "] GetTargetApi=>" + api)
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
func getRandTarget(apiUrls []string) string {
	valLen := len(apiUrls)
	if valLen <= 0 {
		return ""
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
	apiUrl := h.GetID()
	//add apiUrl into alive urls
	if !slicex.Exists(apiInfo.AliveApiUrls, apiUrl){
		apiInfo.AliveApiUrls = append(apiInfo.AliveApiUrls, apiUrl)
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
	apiUrl := h.GetID()
	//add apiUrl into alive urls
	if index := slicex.FindIndex(apiInfo.AliveApiUrls, apiUrl);index != -1{
		apiInfo.AliveApiUrls = append(apiInfo.AliveApiUrls[0:index], apiInfo.AliveApiUrls[index+1:len(apiInfo.AliveApiUrls)]...)
	}
}
