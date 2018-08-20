package models

//网关Api信息
type GatewayApiInfo struct {
	//Api Unique ID
	ApiID int
	//ApiModule
	ApiModule string
	//ApiKey
	ApiKey string
	//API Type
	//Now support ApiType_Balance，ApiType_Group
	ApiType int
	//服务Host类型，默认为录入方式
	ServiceHostType int
	//服务发现注册的服务名
	ServiceDiscoveryName string
	//Api Version, like 1,1.1,1.2
	ApiVersion string
	//Api对应的真实Url
	ApiUrl string
	//Target API
	TargetApi []*TargetApiInfo
	//Alive real target urls, used to do balance
	//auto init, no date in storage
	AliveApiUrls []string
	//http方法，暂时支持Get、Post
	HttpMethod string
	//Api状态 0初始化，100有效，-100暂停
	Status int
	//验证类型：0:不验证；1:MD5验证
	ValidateType int
	//是否返回原始响应字符串，默认为否
	RawResponseFlag bool
	//受限IP，如果为空表示不限制
	ValidIP string
	//受限IP列表，通过ValidIP解析得到
	ValidIPs []string
}

//网关目标API定义
type TargetApiInfo struct {
	TargetKey string
	TargetUrl string
	Weight    int
	//Api状态 0初始化，100有效，-100暂停
	Status int
}

//组合模式下，返回结构定义
type TargetApiResult struct {
	ApiKey       string
	RetCode      int
	RetMsg       string
	IntervalTime int64
	Message      interface{}
}
