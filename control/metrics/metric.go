// metric
package metrics

import (
	"TechPlat/apigateway/config"
	"TechPlat/apigateway/const/log"
	"TechPlat/apigateway/framework/http"
	"TechPlat/apigateway/framework/log"
	"encoding/json"
	"strconv"
	"sync"
	"time"
	"github.com/devfeel/polaris/control/count"
)

var (
	RequestCountMap       map[string]map[string]ApiCountInfo
	Chan_ApiCount         chan ApiCountInfo
	TimeTicker_RequestMap *time.Ticker
	mutex                 *sync.RWMutex
	ServerCounter         count.Counter
)

//Api计数信息
type ApiCountInfo struct {
	AppID      string
	ApiID      string
	ApiModule  string
	ApiName    string
	ApiVersion string
	Count      uint
	RetCode    string
}

const (
	MapMinsTimePrex = "200601021504" //分钟级的time模板

)

func init() {
	RequestCountMap = make(map[string]map[string]ApiCountInfo)
	Chan_ApiCount = make(chan ApiCountInfo, 100000)
	TimeTicker_RequestMap = time.NewTicker(60 * time.Second)
	mutex = new(sync.RWMutex)
	ServerCounter = count.NewCounter()
}

//启动Api计数日志
func StartApiCountHandler() {
	go handleApiCountChan()
	go cronDealRequestMap()
}

// AddRequestCount add request count
func AddRequestCount(num int64){
	ServerCounter.Inc(num)
}

//添加Api计数信息
func AddApiCount(appId string, apiId int, apiModule string, apiName string, apiVersion string, count uint, retCode string) {
	AddRequestCount(1)
	countInfo := ApiCountInfo{
		AppID:      appId,
		ApiID:      strconv.Itoa(apiId),
		ApiModule:  apiModule,
		ApiName:    apiName,
		ApiVersion: apiVersion,
		Count:      count,
		RetCode:    retCode,
	}
	Chan_ApiCount <- countInfo
}

//处理ApiCount队列数据
func handleApiCountChan() {
	for {
		apiCount := <-Chan_ApiCount
		apiTimeKey := time.Now().Format(MapMinsTimePrex)

		mutex.RLock()
		countMap, getFlag := RequestCountMap[apiTimeKey]
		mutex.RUnlock()

		if !getFlag {
			countMap = make(map[string]ApiCountInfo)
		}

		apiKey := apiCount.AppID + "_" + apiCount.ApiModule + "_" + apiCount.ApiVersion + "_" + apiCount.ApiName + "_" + apiCount.RetCode
		if v, ok := countMap[apiKey]; !ok {
			countMap[apiKey] = apiCount
		} else {
			v.Count += 1
			countMap[apiKey] = v
		}

		mutex.Lock()
		RequestCountMap[apiTimeKey] = countMap
		mutex.Unlock()
	}
}

//处理分钟级的Map数据，循环入库
func handleRequestMap() {
	t := time.Now()
	durmi, _ := time.ParseDuration("-60s")
	apiTimeKey := t.Add(durmi).Format(MapMinsTimePrex)

	mutex.RLock()
	apiCounts, ok := RequestCountMap[apiTimeKey]
	mutex.RUnlock()

	if ok {
		mutex.Lock()
		delete(RequestCountMap, apiTimeKey)
		mutex.Unlock()
	}

	if ok {
		//异步处理计数信息
		go func() {
			for apiKey, apiCount := range apiCounts {
				//send api count info to httpapi
				sendApiCountLog(apiCount, apiTimeKey)
				jsonb, _ := json.Marshal(apiCount)
				logger.Debug(apiTimeKey+" => "+apiKey+":"+string(jsonb), logdefine.LogTarget_ApiCount)
			}
		}()
	}

}

//发送ApiCount日志
func sendApiCountLog(apiCount ApiCountInfo, countTime string) {
	apiCountLogUrl := config.CurrentConfig.AppSetting.CountLogApi
	if apiCountLogUrl != "" {
		apiCountLogUrl = apiCountLogUrl +
			"ApiID=" + apiCount.ApiID +
			"&AppID=" + apiCount.AppID +
			"&ApiModule=" + apiCount.ApiModule +
			"&ApiName=" + apiCount.ApiName +
			"&ApiVersion=" + apiCount.ApiVersion +
			"&Count=" + strconv.Itoa(int(apiCount.Count)) +
			"&RetCode=" + apiCount.RetCode +
			"&CountTime=" + countTime
		_, _, _, err := httputil.HttpGet(apiCountLogUrl)
		if err != nil {
			//TODO:file? or email?
		}
	}
}

//定时处理请求Map
func cronDealRequestMap() {
	for {
		select {
		case <-TimeTicker_RequestMap.C:
			handleRequestMap()
		}
	}
}
