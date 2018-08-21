package config

import (
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"time"
	"github.com/devfeel/polaris/models"
	"github.com/devfeel/polaris/util/logx"
	"github.com/devfeel/polaris/const"
	"github.com/devfeel/polaris/util/consul"
	"github.com/devfeel/polaris/cache"
	"errors"
	"io/ioutil"
	"encoding/xml"
	"os"
	"fmt"
	"github.com/devfeel/mapper"
)

const (
	DefaultApiKey  = "default"
	DefaultTestApp = "0"
	ApiIpSplitChar = ","
	DefaultMaxCacheTime = 5
)

var (
	configCacheTime int
	apiMap 		map[string]*models.GatewayApiInfo
	apiMutex	*sync.RWMutex
	appMap 		map[string]*models.AppInfo
	appMutex	*sync.RWMutex
	relationMap	map[string]*models.Relation
	relationMutex	*sync.RWMutex


	LoadConfigTime 		time.Time
	CurrentConfig    	*ProxyConfig
	CurrentBaseDir   	string
	CurrentConfigFile 	string
	innerLogger      	logger.Logger
)

func init() {
	apiMutex = new(sync.RWMutex)
	appMutex = new(sync.RWMutex)
	relationMutex = new(sync.RWMutex)

	innerLogger = logger.InnerLogger
	configCacheTime = DefaultMaxCacheTime
}

func SetBaseDir(baseDir string) {
	CurrentBaseDir = baseDir
}

// InitConfig init config from config file and redis
func InitConfig(configFile string) *ProxyConfig {
	CurrentConfigFile = configFile

	err := initConfig()
	if err != nil{
		fmt.Println("InitConfig err", err)
		os.Exit(1)
	}

	//启动定时重置App、Api信息
	go jobReLoadConfigInfo()

	return CurrentConfig
}

func initConfig() error{
	content, err := ioutil.ReadFile(CurrentConfigFile)
	if err != nil {
		innerLogger.Warn("ProxyConfig::initConfig parse config [" + CurrentConfigFile + "] read error - " + err.Error())
		return err
	}

	result := &ProxyConfig{}
	err = xml.Unmarshal(content, result)
	if err != nil {
		innerLogger.Warn("ProxyConfig::initConfig parse config [" + CurrentConfigFile + "] Unmarshal error - " + err.Error())
		return err
	}
	CurrentConfig = result

	err = loadRedisGlobalConfig()
	if err != nil {
		innerLogger.Warn("ProxyConfig::initConfig loadRedisGlobalConfig error - " + err.Error())
		return err
	}

	//load app & api info
	err = loadRedisAppApiInfo()
	if err != nil {
		innerLogger.Warn("ProxyConfig::initConfig loadRedisAppApiInfo error - " + err.Error())
		return err
	}

	if CurrentConfig.Global.ConfigCacheMins > 0{
		configCacheTime = CurrentConfig.Global.ConfigCacheMins
	}
	return nil
}

func loadRedisGlobalConfig() error{
	innerLogger.Debug("ProxyConfig::initAppMap begin")
	//load data from redis
	redisClient := getRedisCache()
	confJson, err := redisClient.GetString(_const.Redis_Key_GlobalConfig)
	if err != nil {
		innerLogger.Error("ProxyConfig::loadRedisGlobalConfig:GetRedis error: " + err.Error())
		return err
	}

	if confJson == ""{
		innerLogger.Info("ProxyConfig::loadRedisGlobalConfig no have data in redis")
	}else{
		conf := &models.GlobalConfig{}
		errUnmarshal := json.Unmarshal([]byte(confJson), conf)
		if errUnmarshal != nil {
			innerLogger.Error("ProxyConfig::loadRedisGlobalConfig:json.Unmarshal error: " + err.Error())
			return err
		}

		if conf != nil{
			mapper.AutoMapper(conf, &CurrentConfig.Global)
		}
	}

	innerLogger.Debug("ProxyConfig::loadRedisGlobalConfig end")
	return nil
}

func loadRedisAppMap() error{
	innerLogger.Debug("ProxyConfig::initAppMap begin")
	//load data from redis
	redisClient := getRedisCache()
	apps, err := redisClient.HGetAll(_const.Redis_Key_AppMap)
	if err != nil {
		innerLogger.Error("ProxyConfig::initAppMap:GetRedis error: " + err.Error())
		return err
	}

	tmpMap := make(map[string]*models.AppInfo)
	for _, v := range apps {
		app := &models.AppInfo{}
		errUnmarshal := json.Unmarshal([]byte(v), app)
		if errUnmarshal != nil {
			innerLogger.Error("ProxyConfig::initAppMap:json.Unmarshal error: " + err.Error())
		}
		tmpMap[app.AppID] = app
	}

	checkDataLen := 0
	if CurrentConfig.Global.UseDefaultTestApp{
		checkDataLen = 1
		app := &models.AppInfo{
			AppID:   DefaultTestApp,
			AppName: "PolarisDefaultApp",
			AppKey:  "",
			Status:  0,
		}
		tmpMap[app.AppID] = app
	}


	//if load data not exists, no update memory config
	if len(tmpMap) > checkDataLen{
		appMutex.Lock()
		defer appMutex.Unlock()
		appMap = tmpMap
	}

	innerLogger.Debug("ProxyConfig::initAppMap end => " + strconv.Itoa(len(appMap)) + " records")
	return nil
}

func loadRedisApiMap() error{
	innerLogger.Debug("ProxyConfig::initApiMap begin")
	//load data from redis
	redisClient := getRedisCache()
	apis, err := redisClient.HGetAll(_const.Redis_Key_ApiMap)
	if err != nil {
		innerLogger.Error("ProxyConfig::initApiMap:GetRedis error : " + err.Error())
		return err
	}

	tmpMap := make(map[string]*models.GatewayApiInfo)
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

		//load data from consul
		if api.ServiceHostType == _const.ServiceHostType_Consul{
			if api.ApiType == _const.ApiType_Group{
				innerLogger.Error("ProxyConfig::initApiMap group mode not support consul service mode")
				continue
			}
			if !CurrentConfig.Consul.IsUse{
				innerLogger.Error("ProxyConfig::initApiMap Api need consul service but current gateway config not use consul set")
				continue
			}
			if api.ServiceDiscoveryName == ""{
				innerLogger.Error("ProxyConfig::initApiMap Api need consul service but current api service name config is empty")
				continue
			}else{
				tag := ""
				services, err := consul.FindService(CurrentConfig.Consul.ServerUrl, api.ServiceDiscoveryName, tag)
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

	//if load data not exists, no update memory config
	checkDataLen := 0
	if len(tmpMap) > checkDataLen{
		apiMutex.Lock()
		defer apiMutex.Unlock()
		apiMap = tmpMap
	}

	//create alive targets for balance
	for _, v:=range apiMap{
		if v.ApiType == _const.ApiType_Balance {
			v.AliveTargetApis = []*models.TargetApiInfo{}
			for _, api := range v.TargetApis {
				v.AliveTargetApis = append(v.AliveTargetApis, api)
			}
		}
	}

	innerLogger.Debug("ProxyConfig::initApiMap end => " + strconv.Itoa(len(apiMap)) + " records")
	return nil
}

func loadRedisRelationMap() error{
	innerLogger.Debug("ProxyConfig::initRelationMap begin")
	//load data from redis
	redisClient := getRedisCache()
	relations, err := redisClient.HGetAll(_const.Redis_Key_Relation)
	if err != nil {
		innerLogger.Error("ProxyConfig::initRelationMap:GetRedis error: " + err.Error())
		return err
	}

	tmpMap := make(map[string]*models.Relation)
	for k, v := range relations {
		relation :=&models.Relation{}
		errUnmarshal := json.Unmarshal([]byte(v), relation)
		if errUnmarshal != nil {
			innerLogger.Error("ProxyConfig::initRelationMap:json.Unmarshal error: " + err.Error())
		}
		tmpMap[k] = relation
	}

	//if load data not exists, no update memory config
	checkDataLen := 0
	if len(tmpMap) > checkDataLen{
		relationMutex.Lock()
		defer relationMutex.Unlock()
		relationMap = tmpMap
	}
	innerLogger.Debug("ProxyConfig::initRelationMap end => " + strconv.Itoa(len(relationMap)) + " records")
	return nil
}

// loadRedisAppApiInfo
// 1.load app info from redis
// 2.load api info from redis
// 3.load relations in app and api from redis
// 4.update LastConfigTime
func loadRedisAppApiInfo() error{

	if CurrentConfig.Redis.ServerUrl == ""{
		return errors.New("no redis server config")
	}

	//load app list from redis
	err := loadRedisAppMap()
	if err != nil{
		return err
	}

	//load api list from redis
	err = loadRedisApiMap()
	if err != nil{
		return err
	}
	//load the relations in app and api from redis
	err = loadRedisRelationMap()
	if err != nil{
		return err
	}

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
	return nil
}

func getRedisCache() cache.RedisCache{
	return cache.GetRedisCache(CurrentConfig.Redis.ServerUrl, CurrentConfig.Redis.BackupServerUrl, CurrentConfig.Redis.MaxIdle, CurrentConfig.Redis.MaxActive)
}

// job_ReLoadConfigInfo
func jobReLoadConfigInfo() {
	ticker := time.NewTicker(time.Minute * time.Duration(configCacheTime))
	for {
		select {
		case <-ticker.C:
			innerLogger.Debug("ProxyConfig::CronTask_ResetAppApiInfo begin")
			initConfig()
			innerLogger.Debug("ProxyConfig::CronTask_ResetAppApiInfo end")
		}
	}
}

// GetAppInfo get app with appID
func GetAppInfo(appID string) (appInfo *models.AppInfo, ok bool) {
	appMutex.RLock()
	v, mok := appMap[appID]
	appMutex.RUnlock()

	ok = mok
	appInfo = v
	return
}

// GetAppList get all app map
func GetAppList() map[string]*models.AppInfo {
	return appMap
}

// GetApiList get all api map
func GetApiList() map[string]*models.GatewayApiInfo {
	return apiMap
}

// GetRelationList get all relation map
func GetRelationList() map[string]*models.Relation {
	return relationMap
}

// GetApiInfo get api with apiKey
func GetApiInfo(apiKey string) (apiInfo *models.GatewayApiInfo, ok bool) {
	apiMutex.RLock()
	v, mok := apiMap[apiKey]
	apiMutex.RUnlock()

	ok = mok
	apiInfo = v
	return
}

// CheckRelation check relation in app and api
func CheckRelation(appId string, apiId string) (ok bool) {
	//if use default app, have all permission
	if appId == DefaultTestApp {
		return true
	}
	mapKey := appId + "." + apiId
	relationMutex.RLock()
	relation, mok := relationMap[mapKey]
	relationMutex.RUnlock()

	isUse := mok
	if mok {
		isUse = relation.IsUse
	}
	return isUse
}