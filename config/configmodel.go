// ProxyConfigModel
package config

import (
	"encoding/xml"
)

//ProxyConfig config interface
type ProxyConfig struct {
	XMLName    	xml.Name   	`xml:"Config"`
	Server 		Server 		`xml:"Server"`
	AppSetting 	AppSetting 	`xml:"AppSetting"`
	ConsulSet  	ConsulSet 	`xml:"Consul"`
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