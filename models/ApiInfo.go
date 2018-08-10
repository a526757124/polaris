package models

//网关Api信息
type GatewayApiInfo struct {
	//Api编号
	ApiID int
	//ApiModule
	ApiModule string
	//ApiKey，用于定位
	ApiKey string
	//API类型，一般表示为组合或者负载类型
	//参考ApiType_Balance，ApiType_Group
	ApiType int

	//服务Host类型，默认为录入方式
	ServiceHostType int
	//服务发现注册的服务名
	ServiceDiscoveryName string


	//Api版本，例如1,1.1,1.2之类，默认1
	ApiVersion string
	//Api对应的真实Url
	ApiUrl string
	//Target API
	TargetApi []*TargetApiInfo
	//Api对应的存活目标Api数组，用于负载
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
