package service

import (
	"errors"
	"fmt"

	"github.com/a526757124/polaris/uiserver/common/page"
	"github.com/a526757124/polaris/uiserver/conn"
	"github.com/a526757124/polaris/uiserver/viewModel"
)

// AppInfoService 应用服务
type AppInfoService struct{}

func (appInfoService *AppInfoService) getKey() string {
	return new(viewModel.AppInfoDto).TableName()
}

func (appInfoService *AppInfoService) Insert(model *viewModel.AppInfoDto) error {
	n, err := conn.GetMysqlClient().Insert("INSERT INTO `Application`(`Name`,`Desc`,`Status`,`DevUser`,`ProductUser`,`Secret`,`IPList`)VALUES(?,?,?,?,?,?,?);", model.AppName, model.AppDesc, model.Status, model.DevUser, model.ProductUser, "", "")
	if err != nil {
		return err
	}
	fmt.Println(n)
	if n == 0 {
		return errors.New("新增失败！")
	}
	return nil
}
func (appInfoService *AppInfoService) Update(model *viewModel.AppInfoDto) error {
	// key := appInfoService.getKey()
	// score := model.AppID
	// jsonByte, err := json.Marshal(model)
	// n, err := conn.GetRedisClient().ZAdd(key, int64(score), string(jsonByte))
	// if err != nil {
	// 	return errors.New("sorted set insert:" + err.Error())
	// }
	// if n <= 0 {
	// 	return errors.New("更新失败！")
	// }
	return nil
}
func (appInfoService *AppInfoService) Delete(model *viewModel.AppInfoDto) error {
	// key := appInfoService.getKey()
	// jsonByte, err := json.Marshal(model)
	// n, err := conn.GetRedisClient().ZRem(key, string(jsonByte))
	// if err != nil {
	// 	return errors.New("sorted set insert:" + err.Error())
	// }
	// if n <= 0 {
	// 	return errors.New("删除失败！")
	// }
	return nil
}

// getlist
func (appInfoService *AppInfoService) GetList(queryParm *viewModel.AppInfoQueryParm) (*page.PageList, error) {
	pageList := page.PageList{}
	results := []*viewModel.AppInfoDto{}
	err := conn.GetMysqlClient().FindList(&results, "SELECT ID,`Name`,`Desc`,`Status` FROM Application limit ?,?", queryParm.GetSkip(), queryParm.GetLimit())
	if err != nil {
		return nil, err
	}
	count, err := conn.GetMysqlClient().Count("SELECT count(1)  FROM Application")
	if err != nil {
		return nil, err
	}
	pageList.TotalCount = count
	pageList.PageData = results
	return &pageList, nil
}
