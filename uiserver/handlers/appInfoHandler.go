// appInfoHandler
package handlers

import (
	"fmt"
	"strings"

	"github.com/a526757124/polaris/uiserver/common"
	"github.com/a526757124/polaris/uiserver/service"
	"github.com/a526757124/polaris/uiserver/viewModel"
	"github.com/devfeel/dotweb"
)

var appInfoService *service.AppInfoService

func init() {
	appInfoService = &service.AppInfoService{}
}

//appinfo handler
type AppInfoHandler struct{}

//add appinfo
func (*AppInfoHandler) Add(ctx dotweb.Context) error {
	if strings.ToUpper(ctx.Request().Method) == "OPTIONS" {
		ctx.WriteStringC(204, nil)
		return nil
	}
	model := viewModel.AppInfoDto{}
	err := ctx.Bind(&model)
	fmt.Println(model)
	if err != nil {
		ctx.WriteJson(common.NewFailResult("parameter bind fail," + err.Error()))
		return err
	}
	err = appInfoService.Insert(&model)
	if err != nil {
		ctx.WriteJson(common.NewCustomFailResult(10001, err.Error()))
		return err
	}
	ctx.WriteJson(common.NewSuccessResult(nil))
	return nil
}

//update appinfo
func (*AppInfoHandler) Update(ctx dotweb.Context) error {
	if strings.ToUpper(ctx.Request().Method) == "OPTIONS" {
		ctx.WriteStringC(204, nil)
		return nil
	}
	model := viewModel.AppInfoDto{}
	err := ctx.Bind(&model)
	fmt.Println(model)
	if err != nil {
		ctx.WriteJson(common.NewFailResult("parameter bind fail," + err.Error()))
		return err
	}
	err = appInfoService.Update(&model)
	if err != nil {
		ctx.WriteJson(common.NewCustomFailResult(10001, err.Error()))
		return err
	}
	ctx.WriteJson(common.NewSuccessResult(nil))
	return nil
}

func (*AppInfoHandler) Delete(ctx dotweb.Context) error {
	if strings.ToUpper(ctx.Request().Method) == "OPTIONS" {
		ctx.WriteStringC(204, nil)
		return nil
	}
	model := viewModel.AppInfoDto{}
	err := ctx.Bind(&model)
	fmt.Println(model)
	if err != nil {
		ctx.WriteJson(common.NewFailResult("parameter bind fail," + err.Error()))
		return err
	}
	err = appInfoService.Delete(&model)
	if err != nil {
		ctx.WriteJson(common.NewCustomFailResult(10001, err.Error()))
		return err
	}
	ctx.WriteJson(common.NewSuccessResult(nil))
	return nil
}

//得到应用列表
func (*AppInfoHandler) GetList(ctx dotweb.Context) error {
	if strings.ToUpper(ctx.Request().Method) == "OPTIONS" {
		ctx.WriteStringC(204, nil)
		return nil
	}
	queryParam := &viewModel.AppInfoQueryParm{}

	// //appInfoDtoArr := &[]viewModel.AppInfoDto{}
	//自动组装参数
	err := ctx.Bind(queryParam)
	if err != nil {
		ctx.WriteJson(common.NewFailResult("parameter bind fail," + err.Error()))
		return err
	}
	list, err := appInfoService.GetList(queryParam)
	if err != nil {
		ctx.WriteJson(common.NewCustomFailResult(10001, err.Error()))
		return err
	}
	ctx.WriteJson(common.NewSuccessResult(list))
	return nil
}
