package balance

import (
	"TechPlat/apigateway/models"
	"strconv"
	"sync"
	"time"
)

var (
	apiErrorInfo map[string]models.RunApiInfo
	mutex        *sync.RWMutex
)

const (
	//最大错误次数
	MaxErrorCount = 3
)

func init() {
	//初始化读写锁
	mutex = new(sync.RWMutex)
}

type BalanceMonitor struct {
}

//设置错误
func (monitor *BalanceMonitor) SetError(apiInfo models.GatewayApiInfo, apiUrl string) {
	apiKey := strconv.Itoa(apiInfo.ApiID) + "_" + apiUrl
	mutex.RLock()
	v, mok := apiErrorInfo[apiKey]
	mutex.RUnlock()
	if !mok {
		runApiInfo := models.RunApiInfo{
			ApiID:            apiInfo.ApiID,
			ApiUrl:           apiUrl,
			TotalErrorCount:  1,
			LastValidateTime: time.Now(),
		}
		mutex.Lock()
		apiErrorInfo[apiKey] = runApiInfo
		mutex.Unlock()
	} else {
		v.TotalErrorCount += 1
		mutex.Lock()
		apiErrorInfo[apiKey] = v
		mutex.Unlock()
	}
	if v.TotalErrorCount >= MaxErrorCount {
		//TODO:从Alive集合移除
	}
}
