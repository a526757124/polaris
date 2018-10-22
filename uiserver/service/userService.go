package service

import (
	"errors"
	"strconv"

	"github.com/a526757124/polaris/uiserver/common/page"
	"github.com/a526757124/polaris/uiserver/conn"
	"github.com/a526757124/polaris/uiserver/models"
	"github.com/a526757124/polaris/uiserver/viewModel"
)

type UserService struct{}

// get model name
func (userService *UserService) getKey() string {
	return new(models.User).TableName()
}

// insert user
func (userService *UserService) Insert(model *models.User) error {
	n, err := conn.GetMysqlClient().Insert("INSERT INTO `User`(`NickName`,`LoginName`,`LoginPwd`,`Status`)VALUES(?,?,?,?);", model.NickName, model.LoginName, model.LoginPwd, model.Status)
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("新增失败！")
	}
	return nil
}

// update user
func (userService *UserService) Update(model *models.User) error {
	n, err := conn.GetMysqlClient().Update("Update `User` SET `NickName`=?,`LoginName`=?,`Status`=? WHERE `ID`=?;", model.NickName, model.LoginName, model.Status, model.ID)
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("修改失败！")
	}
	return nil
}

// update user pwd
func (userService *UserService) UpdatePwd(model *models.User) error {
	n, err := conn.GetMysqlClient().Update("Update `User` SET `LoginPwd`=? WHERE `ID`=?;", model.LoginPwd, model.ID)
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("修改失败！")
	}
	return nil
}

// delete user
func (userService *UserService) Delete(id int64) error {
	n, err := conn.GetMysqlClient().Delete("DELETE FROM `User` WHERE `ID`=?;", id)
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("删除失败！")
	}
	return nil
}

// get user by id
func (userService *UserService) GetUserById(id int64) (*models.User, error) {
	result := models.User{}
	err := conn.GetMysqlClient().FindOne(&result, "SELECT ID,`NickName`,`LoginName`,`LoginPwd`,`Status`,`CreateTime` FROM `User` where ID=?", id)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// get user by lgoin name
func (userService *UserService) GetUserByLoginName(key string) (*models.User, error) {
	result := models.User{}
	err := conn.GetMysqlClient().FindOne(&result, "SELECT ID,`NickName`,`LoginName`,`LoginPwd`,`Status`,`CreateTime` FROM `User` where LoginName=?", key)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// get user list
func (userService *UserService) GetList(queryParm *viewModel.UserQueryParm) (*page.PageList, error) {
	pageList := page.PageList{}
	results := []*models.User{}
	sql := "SELECT ID,`NickName`,`LoginName`,`LoginPwd`,`Status`,`CreateTime` FROM `User`"
	whereSql := " where 1=1 "
	if queryParm.NickName != "" {
		whereSql += " and `NickName` like '%" + queryParm.NickName + "%'"
	}
	sortSql := ""
	if len(queryParm.SortList) > 0 {
		sortSql += " order by "
		for _, v := range queryParm.SortList {
			sortSql += v.SortField + " " + v.SortWay
		}
	}
	limitSql := " limit " + strconv.FormatInt(queryParm.GetSkip(), 10) + "," + strconv.FormatInt(queryParm.GetLimit(), 10)

	sql += whereSql + sortSql + limitSql
	err := conn.GetMysqlClient().FindList(&results, sql)
	if err != nil {
		return nil, err
	}
	count, err := conn.GetMysqlClient().Count("SELECT count(1)  FROM User" + whereSql)
	if err != nil {
		return nil, err
	}
	pageList.TotalCount = count
	pageList.PageData = results
	return &pageList, nil
}

// is exist user
func (userService *UserService) IsExist(id int64) (bool, error) {
	count, err := conn.GetMysqlClient().Count("SELECT count(*) FROM User where ID=?;", id)
	if err != nil {
		return false, err
	}
	return (count > 0), nil
}

// is exist login name
func (userService *UserService) IsExistByLoginName(loginName string) (bool, error) {
	count, err := conn.GetMysqlClient().Count("SELECT count(*) FROM User where LoginName=?;", loginName)
	if err != nil {
		return false, err
	}
	return (count > 0), nil
}
