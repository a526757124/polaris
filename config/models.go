// ProxyConfigModel
package config

import (
	"encoding/xml"
)

//ProxyConfig config interface
type ProxyConfig struct {
	XMLName    	xml.Name   	`xml:"Config"`
	Server 		Server 		`xml:"Server"`
	Global 		Global 		`xml:"Global"`
	Consul  	Consul		`xml:"Consul"`
	Redis      	Redis      	`xml:"Redis"`
}

//基础应用配置
type Global struct {
	CountLogApi				string
	ConfigCacheMins 		int
	UseDefaultTestApp		bool
}

//Consul config
type Consul struct{
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