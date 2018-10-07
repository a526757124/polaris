package service

import (
	"encoding/json"
	"errors"

	"github.com/a526757124/polaris/uiserver/common/page"
	"github.com/a526757124/polaris/uiserver/conn"
	"github.com/a526757124/polaris/uiserver/repository"
	"github.com/a526757124/polaris/uiserver/viewModel"
)

type APIGroupService struct{}

func (apiGroupService *APIGroupService) getKey() string {
	return new(viewModel.APIGroupDto).TableName()
}

func (apiGroupService *APIGroupService) Insert(model *viewModel.APIGroupDto) error {
	key := apiGroupService.getKey()
	score, err := conn.GetRedisClient().INCR(key + "incr")
	if err != nil {
		return errors.New("sorted set incr:" + err.Error())
	}
	model.GroupID = int64(score)
	jsonByte, err := json.Marshal(model)
	n, err := conn.GetRedisClient().ZAdd(key, int64(score), string(jsonByte))
	if err != nil {
		return errors.New("sorted set insert:" + err.Error())
	}
	if n <= 0 {
		return errors.New("新增失败！")
	}
	return nil
}
func (apiGroupService *APIGroupService) Update(model *viewModel.APIGroupDto) error {
	key := apiGroupService.getKey()
	score := model.GroupID
	jsonByte, err := json.Marshal(model)
	n, err := conn.GetRedisClient().ZAdd(key, int64(score), string(jsonByte))
	if err != nil {
		return errors.New("sorted set insert:" + err.Error())
	}
	if n <= 0 {
		return errors.New("更新失败！")
	}
	return nil
}
func (apiGroupService *APIGroupService) Delete(model *viewModel.APIGroupDto) error {
	key := apiGroupService.getKey()
	jsonByte, err := json.Marshal(model)
	n, err := conn.GetRedisClient().ZRem(key, string(jsonByte))
	if err != nil {
		return errors.New("sorted set insert:" + err.Error())
	}
	if n <= 0 {
		return errors.New("删除失败！")
	}
	return nil
}

// getlist
func (apiGroupService *APIGroupService) GetList(queryParm *viewModel.APIGroupQueryParm) (*page.PageList, error) {
	key := apiGroupService.getKey()
	pageList := page.PageList{}
	//err := repository.GetDataBase().FindByPage(&appInfoDtoArr, queryParm.GetSkip(), queryParm.GetLimit(), "app", queryParm.AppName, queryParm.SortList)
	valueArr, err := conn.GetRedisClient().ZRange(key, queryParm.GetSkip(), (queryParm.GetSkip()+queryParm.GetLimit())-1)
	if err != nil {
		return nil, errors.New("sorted set zrange:" + err.Error())
	}
	var dtoArr []*viewModel.APIGroupDto
	for i := 0; i < len(valueArr); i++ {

		josnStr := valueArr[i]
		jsonByte := []byte(josnStr)
		dto := &viewModel.APIGroupDto{}
		json.Unmarshal(jsonByte, dto)
		dtoArr = append(dtoArr, dto)
	}
	if err != nil {
		return nil, err
	}
	count, err := repository.GetDataBase().Count(key)
	if err != nil {
		return nil, err
	}
	pageList.TotalCount = count
	pageList.PageData = dtoArr
	return &pageList, nil
}
