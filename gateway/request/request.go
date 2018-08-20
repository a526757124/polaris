package request

import (
	"strings"
	"errors"
	"github.com/devfeel/polaris/const"
	"github.com/devfeel/polaris/util/httpx"
	. "github.com/devfeel/polaris/gateway/models"
	"github.com/devfeel/polaris/util/rpcx"
)

// DoRequestTarget do request match HttpGet\HttpPost\JsonRpc
func DoRequestTarget(apiCtx *ApiContext)(body, contentType string, intervalTime int64, err error){
	if apiCtx.RealTargetApi.CallName == _const.CallMethod_HttpGet{
		return doHttpGet(apiCtx)
	}else if apiCtx.RealTargetApi.CallName == _const.CallMethod_HttpPost {
		return doHttpPost(apiCtx)
	}else if apiCtx.RealTargetApi.CallName == _const.CallMethod_JsonRPC {
		return doJsonRpc(apiCtx)
	}else{
		return "", "", 0, errors.New("no match Call Method")
	}
}

func doHttpGet(apiCtx *ApiContext)(body, contentType string, intervalTime int64, err error){
	realApiUrl := combineApiUrl(apiCtx.RealTargetApi.TargetUrl, apiCtx.Query)
	body, contentType, intervalTime, err = httpx.HttpGet(realApiUrl)
	return
}

func doHttpPost(apiCtx *ApiContext)(body, contentType string, intervalTime int64, err error){
	realApiUrl := combineApiUrl(apiCtx.RealTargetApi.TargetUrl, apiCtx.Query)
	body, contentType, intervalTime, err = httpx.HttpPost(realApiUrl, string(apiCtx.PostBody), apiCtx.ContentType)
	return
}

func doJsonRpc(apiCtx *ApiContext)(body, contentType string, intervalTime int64, err error){
	var rpcBody []byte
	rpcBody, intervalTime, err = rpcx.CallJsonRPC(apiCtx.RealTargetApi.TargetUrl, apiCtx.RealTargetApi.CallName, apiCtx.PostBody)
	if err != nil{
		body = string(rpcBody)
	}
	contentType = apiCtx.ContentType
	return
}

// combineApiUrl combine api url and query string
func combineApiUrl(targetApiUrl, queryString string) string{
	//处理参数拼接，考虑是否匹配?和&符号的情况
	if strings.Contains(targetApiUrl, "?") {
		if !strings.HasSuffix(targetApiUrl, "&") && !strings.HasSuffix(targetApiUrl, "?") {
			targetApiUrl = targetApiUrl + "&" + queryString
		} else {
			targetApiUrl = targetApiUrl + queryString
		}

	} else {
		targetApiUrl = targetApiUrl + "?" + queryString
	}
	return targetApiUrl
}