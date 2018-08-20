package config

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"github.com/devfeel/polaris/models"
	"github.com/devfeel/polaris/util/logx"
	"github.com/devfeel/polaris/util/redisx"
	"github.com/devfeel/polaris/const"
	"github.com/devfeel/polaris/util/consul"
)

//默认ApiKey，当指定ApiKey无对应ApiUrl时，返回该项
const (
	DefaultApiKey  = "default"
	ApiSplitChar   = "^$^"
	ApiIpSplitChar = ","
	//最大缓存时间，单位是分钟
	Default_MaxCacheTime = 5
)

var (
	//app、api等缓存时间，单位为分钟
	configCacheTime int
	//Api列表
	apiMap map[string]*models.GatewayApiInfo
	//App列表
	appMap map[string]*models.AppInfo
	//AppApi关系列表
	relationMap      map[string]*models.Relation
	LoadConfigTime time.Time //最后一次加载APIMap的时间
	CurrentConfig    *ProxyConfig
	CurrentBaseDir   string
	allowIPMap       map[string]int
	innerLogger      logger.Logger

	mutex        *sync.RWMutex
	allowIpMutex *sync.RWMutex
)

func init() {
	//初始化读写锁
	mutex = new(sync.RWMutex)
	allowIpMutex = new(sync.RWMutex)
	innerLogger = logger.InnerLogger
	configCacheTime = Default_MaxCacheTime
}

func SetBaseDir(baseDir string) {
	CurrentBaseDir = baseDir
}

//初始化配置信息
func InitConfig(configFile string) *ProxyConfig {
	content, err := ioutil.ReadFile(configFile)
	if err != nil {
		innerLogger.Warn("ProxyConfig::InitConfig 配置文件[" + configFile + "]无法解析 - " + err.Error())
		os.Exit(1)
	}

	result := &ProxyConfig{}
	err = xml.Unmarshal(content, result)
	if err != nil {
		innerLogger.Warn("ProxyConfig::InitConfig 配置文件[" + configFile + "]解析失败 - " + err.Error())
		os.Exit(1)
	}
	CurrentConfig = result

	//初始化允许访问的IP字典
	allowIPMap = make(map[string]int)
	for _, v := range CurrentConfig.AllowIPs {
		allowIPMap[v] = 1
	}

	//初始化App、Api信息
	resetAppApiInfo()

	if CurrentConfig.AppSetting.ConfigCacheMins > 0{
		configCacheTime = CurrentConfig.AppSetting.ConfigCacheMins
	}

	//启动定时重置App、Api信息
	go CronTask_ResetAppApiInfo()

	return CurrentConfig
}

//初始化App信息
func initAppMap() {
	innerLogger.Debug("ProxyConfig::initAppMap begin")
	//从redis获取数据
	redisClient := redisx.GetRedisClient(CurrentConfig.Redis.ServerIP, CurrentConfig.Redis.MaxIdle, CurrentConfig.Redis.MaxActive)
	apps, err := redisClient.HGetAll(_const.Redis_Key_AppMap)
	if err != nil {
		innerLogger.Error("ProxyConfig::initAppMap:redisClient.HGetAll error: " + err.Error())
		return
	}

	tmpMap := make(map[string]*models.AppInfo)
	//处理Redis配置
	for _, v := range apps {
		app := &models.AppInfo{}
		errUnmarshal := json.Unmarshal([]byte(v), app)
		if errUnmarshal != nil {
			innerLogger.Error("ProxyConfig::initAppMap:json.Unmarshal error: " + err.Error())
		}
		tmpMap[strconv.Itoa(app.AppID)] = app
	}

	//特殊处理
	app := &models.AppInfo{
		AppID:   10000,
		AppName: "基础应用",
		AppKey:  "",
		Status:  0,
	}
	tmpMap[strconv.Itoa(app.AppID)] = app

	//处理极端数据情况
	if len(tmpMap) > 1{
		mutex.Lock()
		defer mutex.Unlock()
		appMap = tmpMap
	}
	innerLogger.Debug("ProxyConfig::initAppMap end => " + strconv.Itoa(len(appMap)) + " records")
}

//初始化ApiMap
func initApiMap() {
	innerLogger.Debug("ProxyConfig::initApiMap begin")
	//从redis获取数据
	redisClient := redisx.GetRedisClient(CurrentConfig.Redis.ServerIP, CurrentConfig.Redis.MaxIdle, CurrentConfig.Redis.MaxActive)

	apis, err := redisClient.HGetAll(_const.Redis_Key_ApiMap)
	if err != nil {
		innerLogger.Error("ProxyConfig::initApiMap:redisClient.HGetAll error : " + err.Error())
		return
	}

	tmpMap := make(map[string]*models.GatewayApiInfo)
	//处理Redis配置
	for _, v := range apis {
		api := &models.GatewayApiInfo{}
		errUnmarshal := json.Unmarshal([]byte(v), api)
		if errUnmarshal != nil {
			innerLogger.Error("ProxyConfig::initApiMap:json.Unmarshal error : " + err.Error())
			break
		}
		if api.ValidIP != "" {
			api.ValidIPs = strings.Split(api.ValidIP, ApiIpSplitChar)
		}

		//Consul处理
		if api.ServiceHostType == _const.ServiceHostType_Consul{
			if api.ApiType == _const.ApiType_Group{
				innerLogger.Error("ProxyConfig::initApiMap group mode not support consul service mode")
				continue
			}
			if !CurrentConfig.ConsulSet.IsUse{
				innerLogger.Error("ProxyConfig::initApiMap Api need consul service but current gateway config not use consul set")
				continue
			}
			if api.ServiceDiscoveryName == ""{
				innerLogger.Error("ProxyConfig::initApiMap Api need consul service but current api service name config is empty")
				continue
			}else{
				tag := ""
				services, err := consul.FindService(CurrentConfig.ConsulSet.ServerUrl, api.ServiceDiscoveryName, tag)
				if err != nil{
					innerLogger.Error("ProxyConfig::initApiMap Api need consul service but load services error" + err.Error())
					continue
				}
				if len(services) <= 0{
					innerLogger.Error("ProxyConfig::initApiMap api target api service host info is empty")
					continue
				}
				if len(api.TargetApis) <= 0{
					innerLogger.Error("ProxyConfig::initApiMap api target api path is empty")
					continue
				}
				//默认取第一个配置项
				pathUrl := api.TargetApis[0].TargetUrl
				api.TargetApis = []*models.TargetApiInfo{}
				//init TargetApiInfo
				weight := 100 / len(services)
				for index, v:=range services{
					api.TargetApis = append(api.TargetApis, &models.TargetApiInfo{
						TargetKey:strconv.Itoa(index),
						TargetUrl:v.Address + ":" + strconv.Itoa(v.Port) + pathUrl,
						Weight:weight,
						Status:_const.ApiStatus_Normal,
					})
				}
				innerLogger.Debug("ProxyConfig::initApiMap Api init consul services" + strconv.Itoa(len(services)))
			}
		}

		tmpMap[api.ApiModule+"/"+api.ApiKey+"/"+api.ApiVersion] = api
	}

	//处理配置文件配置
	//配置文件不支持group模式
	for _, api := range CurrentConfig.LocalApis {
		gateApi := &models.GatewayApiInfo{
			ApiKey:       api.ApiKey,
			ApiModule:    api.Module,
			ApiVersion:   api.ApiVersion,
			Status:       api.Status,
			ValidateType: api.ValidateType,
			ValidIP:      api.ValidIP,
		}
		gateApi.TargetApis = []*models.TargetApiInfo{
			&models.TargetApiInfo{TargetKey:"local", TargetUrl:api.ApiUrl, CallName:api.CallName, CallMethod:api.CallName},
		}
		if gateApi.ValidIP != "" {
			gateApi.ValidIPs = strings.Split(gateApi.ValidIP, ApiIpSplitChar)
		}
		tmpMap[api.Module+"/"+api.ApiKey+"/"+api.ApiVersion] = gateApi
	}

	//处理极端数据情况
	if len(tmpMap) > 0{
		mutex.Lock()
		defer mutex.Unlock()
		apiMap = tmpMap
	}

	//create alive urls for balance
	for _, v:=range apiMap{
		if v.ApiType == _const.ApiType_Balance {
			v.AliveTargetApis = []*models.TargetApiInfo{}
			for _, api := range v.TargetApis {
				v.AliveTargetApis = append(v.AliveTargetApis, api)
			}
		}
	}

	innerLogger.Debug("ProxyConfig::initApiMap end => " + strconv.Itoa(len(apiMap)) + " records")

}

//初始化AppApiRelation信息
func initAppApiRelationMap() {
	innerLogger.Debug("ProxyConfig::initAppApiRelationMap begin")
	//从redis获取数据
	redisClient := redisx.GetRedisClient(CurrentConfig.Redis.ServerIP, CurrentConfig.Redis.MaxIdle, CurrentConfig.Redis.MaxActive)
	relations, err := redisClient.HGetAll(_const.Redis_Key_AppApiRelation)
	if err != nil {
		innerLogger.Error("ProxyConfig::initAppApiRelationMap:redisClient.HGetAll error: " + err.Error())
		return
	}

	tmpMap := make(map[string]*models.Relation)
	//处理Redis配置
	for k, v := range relations {
		relation :=&models.Relation{}
		errUnmarshal := json.Unmarshal([]byte(v), relation)
		if errUnmarshal != nil {
			innerLogger.Error("ProxyConfig::initAppApiRelationMap:json.Unmarshal error: " + err.Error())
		}
		tmpMap[k] = relation
	}

	//处理极端数据情况
	if len(tmpMap) > 0{
		mutex.Lock()
		defer mutex.Unlock()
		relationMap = tmpMap
	}
	innerLogger.Debug("ProxyConfig::initAppApiRelationMap end => " + strconv.Itoa(len(relationMap)) + " records")
}

/*重置App、Api信息
* Author: Panxinming
* LastUpdateTime: 2016-05-30 18:00
* 1）从Redis中Hashset结构获取App配置信息，置入本地Map
* 2）从Redis中Hashset结构获取Api配置信息，置入本地Map
* 3）更新最后加载Api时间属性，用于后续字典有效期判断
 */
func resetAppApiInfo() {

	defer func() {
		var errmsg string
		if err := recover(); err != nil {
			errmsg = fmt.Sprintln(err)
			os.Stdout.Write([]byte("proxyconfig:resetAppApiInfo error! => " + errmsg))
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, true)
			innerLogger.Error("proxyconfig:resetAppApiInfo error! => " + errmsg + " => " + string(buf[:n]))
		}
	}()

	//init app list from config
	initAppMap()
	//init api list from config
	initApiMap()
	//init the relations in app and api
	initAppApiRelationMap()

	LoadConfigTime = time.Now()

	for _, api := range apiMap {
		jsons, _ := json.Marshal(api)
		innerLogger.Debug("ProxyConfig::LoadAppApiInfo ConfigApi=>" + string(jsons))
	}

	for _, app := range appMap {
		jsons, _ := json.Marshal(app)
		innerLogger.Debug("ProxyConfig::LoadAppApiInfo ConfigApp=>" + string(jsons))
	}

	for _, relation := range relationMap {
		jsons, _ := json.Marshal(relation)
		innerLogger.Debug("ProxyConfig::LoadAppApiInfo ConfigAppApiRelation=>" + string(jsons))
	}
}

//计划任务-重置App、Api信息
//间隔时间依据MaxCacheTime设置
func CronTask_ResetAppApiInfo() {
	TimeTicker_Task := time.NewTicker(time.Minute * time.Duration(configCacheTime))
	for {
		select {
		case <-TimeTicker_Task.C:
			innerLogger.Debug("ProxyConfig::CronTask_ResetAppApiInfo begin")
			resetAppApiInfo()
			innerLogger.Debug("ProxyConfig::CronTask_ResetAppApiInfo end")
		}
	}
}

//根据AppID获取对应的AppInfo
func GetAppInfo(appID string) (appInfo *models.AppInfo, ok bool) {
	mutex.RLock()
	v, mok := appMap[appID]
	mutex.RUnlock()

	ok = mok
	appInfo = v
	return
}

func GetAppList() map[string]*models.AppInfo {
	return appMap
}

func GetApiList() map[string]*models.GatewayApiInfo {
	return apiMap
}

func GetRelationList() map[string]*models.Relation {
	return relationMap
}

//根据apiKey获取对应的ApiUrl
func GetApiInfo(apiKey string) (apiInfo *models.GatewayApiInfo, ok bool) {
	mutex.RLock()
	v, mok := apiMap[apiKey]
	mutex.RUnlock()

	ok = mok
	apiInfo = v
	return
}

//检查指定App与Api是否存在权限关系
func CheckAppApiRelation(appId int, apiId int) (ok bool) {
	//特殊的，如果为测试应用，默认放行
	if appId == 10000 {
		return true
	}
	mapKey := strconv.Itoa(appId) + "." + strconv.Itoa(apiId)
	mutex.RLock()
	relation, mok := relationMap[mapKey]
	mutex.RUnlock()

	isUse := mok
	if mok {
		isUse = relation.IsUse
	}
	return isUse
}

//检查指定IP是否在允许列表内
func CheckAllowIP(clientIp string) (ok bool) {
	allowIpMutex.RLock()
	_, mok := allowIPMap[clientIp]
	allowIpMutex.RUnlock()
	return mok
}
