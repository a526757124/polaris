package handlers

import (
	"strconv"
	"strings"

	"github.com/a526757124/polaris/uiserver/common"
	"github.com/a526757124/polaris/uiserver/models"
	"github.com/a526757124/polaris/uiserver/service"
	"github.com/a526757124/polaris/uiserver/viewModel"
	"github.com/devfeel/dotweb"
)

var userService *service.UserService

func init() {
	userService = &service.UserService{}
}

//user handler
type UserHandler struct{}

//add user
func (*UserHandler) Add(ctx dotweb.Context) error {
	if strings.ToUpper(ctx.Request().Method) == "OPTIONS" {
		ctx.WriteStringC(204, nil)
		return nil
	}
	model := models.User{}
	err := ctx.Bind(&model)
	if err != nil {
		ctx.WriteJson(common.NewFailResult("parameter bind fail," + err.Error()))
		return err
	}
	//校验
	//登录名是否存在
	isExist, err := userService.IsExistByLoginName(model.LoginName)
	if err != nil {
		ctx.WriteJson(common.NewCustomFailResult(10001, err.Error()))
		return err
	}
	if isExist {
		ctx.WriteJson(common.NewCustomFailResult(20001, "登录名已存在！"))
		return err
	}
	err = userService.Insert(&model)
	if err != nil {
		ctx.WriteJson(common.NewCustomFailResult(10001, err.Error()))
		return err
	}
	ctx.WriteJson(common.NewSuccessResult(nil))
	return nil
}

//Update user
func (*UserHandler) Update(ctx dotweb.Context) error {
	if strings.ToUpper(ctx.Request().Method) == "OPTIONS" {
		ctx.WriteStringC(204, nil)
		return nil
	}
	model := models.User{}
	err := ctx.Bind(&model)
	if err != nil {
		ctx.WriteJson(common.NewFailResult("parameter bind fail," + err.Error()))
		return err
	}
	//校验
	//登录名是否存在
	isExist, err := userService.IsExistByLoginName(model.LoginName)
	if err != nil {
		ctx.WriteJson(common.NewCustomFailResult(10001, err.Error()))
		return err
	}
	if isExist {
		ctx.WriteJson(common.NewCustomFailResult(20001, "登录名已存在！"))
		return err
	}
	err = userService.Update(&model)
	if err != nil {
		ctx.WriteJson(common.NewCustomFailResult(10001, err.Error()))
		return err
	}
	ctx.WriteJson(common.NewSuccessResult(nil))
	return nil
}

//Update user pwd
func (*UserHandler) UpdatePwd(ctx dotweb.Context) error {
	if strings.ToUpper(ctx.Request().Method) == "OPTIONS" {
		ctx.WriteStringC(204, nil)
		return nil
	}
	model := models.User{}
	err := ctx.Bind(&model)
	if err != nil {
		ctx.WriteJson(common.NewFailResult("parameter bind fail," + err.Error()))
		return err
	}
	err = userService.UpdatePwd(&model)
	if err != nil {
		ctx.WriteJson(common.NewCustomFailResult(10001, err.Error()))
		return err
	}
	ctx.WriteJson(common.NewSuccessResult(nil))
	return nil
}

//Delete user
func (*UserHandler) Delete(ctx dotweb.Context) error {
	if strings.ToUpper(ctx.Request().Method) == "OPTIONS" {
		ctx.WriteStringC(204, nil)
		return nil
	}
	model := models.User{}
	err := ctx.Bind(&model)
	if err != nil {
		ctx.WriteJson(common.NewFailResult("parameter bind fail," + err.Error()))
		return err
	}
	err = userService.Delete(model.ID)
	if err != nil {
		ctx.WriteJson(common.NewCustomFailResult(10001, err.Error()))
		return err
	}
	ctx.WriteJson(common.NewSuccessResult(nil))
	return nil
}

//login user
func (*UserHandler) Login(ctx dotweb.Context) error {
	if strings.ToUpper(ctx.Request().Method) == "OPTIONS" {
		ctx.WriteStringC(204, nil)
		return nil
	}
	model := models.User{}
	err := ctx.Bind(&model)
	if err != nil {
		ctx.WriteJson(common.NewFailResult("parameter bind fail," + err.Error()))
		return err
	}
	user, err := userService.GetUserByLoginName(model.LoginName)
	if err != nil {
		ctx.WriteJson(common.NewCustomFailResult(10001, err.Error()))
		return err
	}
	if user == nil {
		ctx.WriteJson(common.NewCustomFailResult(20001, "登录名不存在！"))
		return nil
	}
	if user.LoginPwd != model.LoginPwd {
		ctx.WriteJson(common.NewCustomFailResult(20001, "登录密码错误！"))
		return nil
	}
	//登录成功后返回token
	loginUser := new(viewModel.LoginUser)
	loginUser.Token = strconv.FormatInt(user.ID, 10)
	ctx.WriteJson(common.NewSuccessResult(loginUser))
	return nil
}

//get login info
func (*UserHandler) GetLoginInfo(ctx dotweb.Context) error {
	if strings.ToUpper(ctx.Request().Method) == "OPTIONS" {
		ctx.WriteStringC(204, nil)
		return nil
	}
	loginUser := new(viewModel.LoginUser)
	err := ctx.Bind(loginUser)
	if err != nil {
		ctx.WriteJson(common.NewFailResult("parameter bind fail," + err.Error()))
		return err
	}
	userID, err := strconv.ParseInt(loginUser.Token, 10, 64)
	if err != nil {
		ctx.WriteJson(common.NewFailResult("parameter convent fail," + err.Error()))
		return err
	}
	user, err := userService.GetUserById(userID)
	if err != nil {
		ctx.WriteJson(common.NewCustomFailResult(10002, "用户登录已过期，请重新登录！"))
		return err
	}
	ctx.WriteJson(common.NewSuccessResult(user))
	return nil
}

//get user list
func (*UserHandler) GetList(ctx dotweb.Context) error {
	if strings.ToUpper(ctx.Request().Method) == "OPTIONS" {
		ctx.WriteStringC(204, nil)
		return nil
	}
	queryParm := viewModel.UserQueryParm{}
	err := ctx.Bind(&queryParm)
	if err != nil {
		ctx.WriteJson(common.NewFailResult("parameter bind fail," + err.Error()))
		return err
	}
	list, err := userService.GetList(&queryParm)
	if err != nil {
		ctx.WriteJson(common.NewCustomFailResult(10001, err.Error()))
		return err
	}
	ctx.WriteJson(common.NewSuccessResult(list))
	return nil
}
