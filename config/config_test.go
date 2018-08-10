package config


import (
	"testing"
	"encoding/json"
	"TechPlat/apigateway/models"
	"fmt"
)

var jsonString = `{"ApiID":149,"ApiModule":"logistics","ApiKey":"oms.checkprodsisbuy","ApiVersion":"1.0","ApiUrl":"http://192.168.8.183:8090/SCMAPI/OMS/CheckProdsIsBuy","HttpMethod":"Get","Status":100,"ValidateType":0,"RawResponseFlag":false,"ApiDese":"查询用户是否购买指定物流包","DevUser":"赵勇","Product":"物流","Operator":"zhangjian2@emoney.cn","WriteTime":"2017-11-28T11:32:14","LastModifyTime":"0001-01-01T00:00:00","ValidIP":"","ApiType":0,"TargetApiUrl":null,"TargetApi":[{"TargetKey":"key1","TargetUrl":"http://www.emoney.cn","Weight":1,"Status":100},{"TargetKey":"key2","TargetUrl":"http://fuwu.emoney.cn","Weight":1,"Status":100}]}`

func Test_ApiJson(t *testing.T){
	fmt.Println(jsonString)
	var api models.GatewayApiInfo
	errUnmarshal := json.Unmarshal([]byte(jsonString), &api)
	t.Log(errUnmarshal, api)
}

