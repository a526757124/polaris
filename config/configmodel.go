// ProxyConfigModel
package config

import (
	"encoding/xml"
)

//代理配置信息
type ProxyConfig struct {
	XMLName    xml.Name   `xml:"proxyconfig"`
	HttpServer HttpServer `xml:"httpserver"`
	LocalApis       []LocalApi      `xml:"apis>api"`
	AllowIPs   []string   `xml:"allowips>ip"`
	AppSetting AppSetting `xml:"appsetting"`
	ConsulSet  ConsulSet `xml:"consul"`
	Redis      Redis      `xml:"redis"`
}

//基础应用配置
type AppSetting struct {
	CountLogApi            string `xml:"countlogapi"`
	ApiCallNumLimitPerMins int    `xml:"apicallnumlimitpermins"`
	ConfigCacheMins int `xml:"configcachemins"`
}

//Consul 配置
type ConsulSet struct{
	IsUse bool `xml:"isuse,attr"`
	ServerUrl string `xml:"serverurl,attr"`
}

//全局配置
type HttpServer struct {
	HttpPort int `xml:"httpport,attr"`
}

//Redis配置
type Redis struct {
	ServerIP string `xml:"serverip,attr"`
	MaxIdle int `xml:"maxidle,attr"`
	MaxActive int `xml:"maxactive,attr"`
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

