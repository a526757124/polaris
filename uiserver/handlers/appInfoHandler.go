// appInfoHandler
package handlers

import (
	"github.com/a526757124/polaris/uiserver/common"
	"github.com/a526757124/polaris/uiserver/viewModel"
	"github.com/a526757124/polaris/util/redisx"
	"github.com/devfeel/dotweb"
	"github.com/devfeel/mapper"
)

var redisClient *redisx.RedisClient

const (
	AppInfoKey = "appInfo"
)

func init() {
	mapper.Register(&viewModel.AppInfoQueryParm{})
	//redisClient = redisx.GetRedisClient()
}

//得到应用列表
func GetList(ctx dotweb.Context) error {
	queryParam := &viewModel.AppInfoQueryParm{}
	appInfoDtoArr := &[]viewModel.AppInfoDto{}
	//自动组装参数
	err := ctx.Bind(&queryParam)
	if err != nil {
		ctx.WriteJson(common.NewFailResult("parameter bind fail," + err.Error()))
		return err
	}
	//取总数
	count, err := redisClient.ZCard(AppInfoKey)
	if err != nil {
		ctx.WriteJson(common.NewCustomFailResult(10001, err.Error()))
		return err
	}
	//当总数为0时直接返回
	if count == 0 {
		ctx.WriteJson(common.NewSuccessResult(appInfoDtoArr))
		return nil
	}
	//取key的所有数量
	resJson, err := redisClient.ZRange(AppInfoKey, int64(0), int64(count))
	if err != nil {
		ctx.WriteJson(common.NewCustomFailResult(10001, err.Error()))
		return err
	}
	ctx.WriteJson(resJson)
	return nil
}
