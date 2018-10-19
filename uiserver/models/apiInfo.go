package models

//api信息
type APIInfo struct {
	ID                   int64
	Name                 string //名称
	Type                 int64  //api类型 1负载 2组合
	Desc                 string //描述
	GroupID              int64  //所属组
	DevUser              string //开发人员
	ServiceHostType      int64  //服务Host类型
	ServiceDiscoveryName string //服务发现注册的服务名
	ValidateType         int64  //校验类型 0:不验证；1:MD5验证
	Version              string //Api版本号
	ApiPath              string //API请求路径
	SupportProtocol      int64  //支持协议 1.http 2.rpc
	ReqMethod            int64  //请求方式 http 协议支持1.get 2.post 3.put 4.delete rpc 协议支持 5.jsonrpc
	TargetApis           string //目标api服务 [{"TargetKey":"","TargetUrl":"","CallMethod":"","CallName":"","Weight":0,"Status":0,"Timeout":0}]
	IsUseMock            bool   //是否启用Mock 0:不启用；1:启用
	MockData             string //mock请求返回数据
	Status               int64  //状态 0:是初始化 1:是有效 -1:是无效
	RawResponseFlag      bool   //是否返回原始响应字符串 0:不返回；1:返回
	ResultType           int64  //返回ContentType 1.json 2.文本 3.二进制 4.xml 5.html
	ResultSample         string //返回结果示例
	FailResultSample     string //失败返回结果示例
	CreateUser           int64  //创建人
	CreateTime           int64  //创建时间
}
