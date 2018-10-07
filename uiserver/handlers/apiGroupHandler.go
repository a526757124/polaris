// APIGroupHandler
package handlers

import (
	"fmt"

	"github.com/a526757124/polaris/uiserver/common"
	"github.com/a526757124/polaris/uiserver/service"
	"github.com/a526757124/polaris/uiserver/viewModel"
	"github.com/devfeel/dotweb"
)

var apiGroupService *service.APIGroupService

func init() {
	apiGroupService = new(service.APIGroupService)
}

//APIGroup handler
type APIGroupHandler struct{}

//add
func (*APIGroupHandler) Add(ctx dotweb.Context) error {
	model := viewModel.APIGroupDto{}
	err := ctx.Bind(&model)
	fmt.Println(model)
	if err != nil {
		ctx.WriteJson(common.NewFailResult("parameter bind fail," + err.Error()))
		return err
	}
	err = apiGroupService.Insert(&model)
	if err != nil {
		ctx.WriteJson(common.NewCustomFailResult(10001, err.Error()))
		return err
	}
	ctx.WriteJson(common.NewSuccessResult(nil))
	return nil
}

//update
func (*APIGroupHandler) Update(ctx dotweb.Context) error {
	model := viewModel.APIGroupDto{}
	err := ctx.Bind(&model)
	fmt.Println(model)
	if err != nil {
		ctx.WriteJson(common.NewFailResult("parameter bind fail," + err.Error()))
		return err
	}
	err = apiGroupService.Update(&model)
	if err != nil {
		ctx.WriteJson(common.NewCustomFailResult(10001, err.Error()))
		return err
	}
	ctx.WriteJson(common.NewSuccessResult(nil))
	return nil
}

func (*APIGroupHandler) Delete(ctx dotweb.Context) error {
	model := viewModel.APIGroupDto{}
	err := ctx.Bind(&model)
	fmt.Println(model)
	if err != nil {
		ctx.WriteJson(common.NewFailResult("parameter bind fail," + err.Error()))
		return err
	}
	err = apiGroupService.Delete(&model)
	if err != nil {
		ctx.WriteJson(common.NewCustomFailResult(10001, err.Error()))
		return err
	}
	ctx.WriteJson(common.NewSuccessResult(nil))
	return nil
}

//得到应用列表
func (*APIGroupHandler) GetList(ctx dotweb.Context) error {
	queryParam := &viewModel.APIGroupQueryParm{}
	// //appInfoDtoArr := &[]viewModel.AppInfoDto{}
	//自动组装参数
	err := ctx.Bind(queryParam)
	if err != nil {
		ctx.WriteJson(common.NewFailResult("parameter bind fail," + err.Error()))
		return err
	}
	list, err := apiGroupService.GetList(queryParam)
	if err != nil {
		ctx.WriteJson(common.NewCustomFailResult(10001, err.Error()))
		return err
	}
	ctx.WriteJson(common.NewSuccessResult(list))
	return nil
}
