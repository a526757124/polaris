// ProxyConfigModel
package config

import (
	"encoding/xml"
)

//代理配置信息
type ProxyConfig struct {
	XMLName    	xml.Name   	`xml:"Config"`
	Server 		Server 		`xml:"Server"`
	LocalApis  	[]LocalApi 	`xml:"Apis>Api"`
	AppSetting 	AppSetting 	`xml:"AppSetting"`
	ConsulSet  	ConsulSet 	`xml:"consul"`
	Redis      	Redis      	`xml:"Redis"`
}

//基础应用配置
type AppSetting struct {
	CountLogApi            string
	ApiCallNumLimitPerMins int
	ConfigCacheMins int
}

//Consul config
type ConsulSet struct{
	IsUse bool
	ServerUrl string
}

//Server server config
type Server struct {
	HttpPort int
	JsonRpcPort int
}

//Redis redis config
type Redis struct {
	ServerUrl string
	BackupServerUrl string
	MaxIdle int
	MaxActive int
}

//Api配置
type LocalApi struct {
	Module       string `xml:"module"`
	ApiKey       string `xml:"apikey"`
	ApiVersion   string `xml:"apiversion"`
	ApiUrl       string `xml:"apiurl"`
	CallMethod   string
	CallName	 string
	Status       int    `xml:"status"`
	ValidateType int    `xml:"validatetype"`
	ValidIP      string `xml:"validip"`
}

