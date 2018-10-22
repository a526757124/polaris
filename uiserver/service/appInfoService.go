package service

import (
	"errors"
	"fmt"

	"github.com/a526757124/polaris/uiserver/common/page"
	"github.com/a526757124/polaris/uiserver/conn"
	"github.com/a526757124/polaris/uiserver/models"
	"github.com/a526757124/polaris/uiserver/viewModel"
)

// AppInfoService 应用服务
type AppInfoService struct{}

func (appInfoService *AppInfoService) getKey() string {
	return new(viewModel.AppInfoDto).TableName()
}

func (appInfoService *AppInfoService) Insert(model *viewModel.AppInfoDto) error {
	n, err := conn.GetMysqlClient().Insert("INSERT INTO `Application`(`Name`,`Desc`,`Status`,`DevUser`,`ProductUser`,`Secret`,`IPList`)VALUES(?,?,?,?,?,?,?);",
		model.AppName, model.AppDesc, model.Status, model.DevUser, model.ProductUser, "", "")
	if err != nil {
		return err
	}
	fmt.Println(n)
	if n == 0 {
		return errors.New("新增失败！")
	}
	return nil
}
func (appInfoService *AppInfoService) Update(model *models.AppInfo) error {
	n, err := conn.GetMysqlClient().Update("Update `Application` SET `Name`=?,`Desc`=?,`Url`=?,`IPList`=?,`DevUser`=?,'ProductUser'=?,`Status`=?,`CreateUser`=? WHERE `ID`=?;",
		model.Name, model.Desc, model.Url, model.IPList, model.DevUser, model.ProductUser, model.Status, model.CreateUser, model.ID)
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("修改失败！")
	}
	return nil
}
func (appInfoService *AppInfoService) Delete(id int64) error {
	n, err := conn.GetMysqlClient().Delete("DELETE FROM `Application` WHERE `ID`=?;", id)
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("删除失败！")
	}
	return nil
}

// get appinfo list
func (appInfoService *AppInfoService) GetList(queryParm *viewModel.AppInfoQueryParm) (*page.PageList, error) {
	pageList := page.PageList{}
	results := []*models.AppInfo{}

	sql := "SELECT ID,`Name`,`Desc`,`Key`,`Url`,`IPList`,`DevUser`,`ProductUser`,`Status`,`CreateUser`,`CreateTime`  FROM `Application`"
	whereSql := " where 1=1 "
	if queryParm.AppName != "" {
		whereSql += " and `AppName` like '%" + queryParm.AppName + "%'"
	}
	sql += whereSql + queryParm.GetSortSql() + queryParm.GetPageSql()
	err := conn.GetMysqlClient().FindList(&results, sql)
	if err != nil {
		return nil, err
	}
	count, err := conn.GetMysqlClient().Count("SELECT count(1)  FROM Application" + whereSql)
	if err != nil {
		return nil, err
	}
	pageList.TotalCount = count
	pageList.PageData = results
	return &pageList, nil
}
