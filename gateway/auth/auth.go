package auth

import (
	"strings"
	"sort"
	"github.com/devfeel/polaris/util/hashx"
	"net/url"
)

// CheckSign check md5 sign with query args and post body
func CheckSign(queryArgs url.Values, postBody string, md5Key string, appEncrypt string) (appVal, gateVal string, isOk bool){
	querys := strings.Split(queryArgs.Encode(), "&")
	sort.Strings(querys)
	querySource := strings.Join(querys, "")
	//add post string
	querySource += postBody
	querySource += md5Key
	encrypt := hashx.MD5(querySource)
	gateVal = strings.ToLower(encrypt)
	appVal = strings.ToLower(appEncrypt)
	//fmt.Println(ctx.Url(), " => checkEncrypt => ", querySource, " || ", gateVal, " || ", appVal)
	return appVal, gateVal, appVal == gateVal
}
