package redis

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/a526757124/polaris/uiserver/conn"
)

//sorted set 存储
type SortedSet struct {
}

//Insert
func (store *SortedSet) Insert(args ...interface{}) (n int64, err error) {
	//conn.GetRedisClient().ZAdd()
	if len(args) == 0 {
		return 0, errors.New("args is nil")
	}
	key, b := args[0].(string)
	if !b {
		return 0, errors.New("one arg is nil")
	}
	o := args[1]
	score, err := conn.GetRedisClient().INCR(key + "incr")
	if err != nil {
		return 0, errors.New("sorted set incr:" + err.Error())
	}
	jsonByte, err := json.Marshal(o)
	n, err = conn.GetRedisClient().ZAdd(key, int64(score), string(jsonByte))
	if err != nil {
		return 0, errors.New("sorted set insert:" + err.Error())
	}
	return n, nil
}

//Update
func (store *SortedSet) Update(args ...interface{}) (n int64, err error) {
	return 0, nil
}

//Delete
func (store *SortedSet) Delete(args ...interface{}) (n int64, err error) {
	return 0, nil
}

//Find
func (store *SortedSet) Find(dest interface{}, args ...interface{}) error {

	return nil
}

//FindOne
func (store *SortedSet) FindOne(dest interface{}, args ...interface{}) error {
	return nil
}

//FindByPage
func (store *SortedSet) FindByPage(dest interface{}, skip, limit int64, args ...interface{}) error {
	if len(args) == 0 {
		return errors.New("args is nil")
	}
	key, b := args[0].(string)
	if !b {
		return errors.New("one arg is nil")
	}
	valueArr, err := conn.GetRedisClient().ZRange(key, skip, (skip+limit)-1)
	if err != nil {
		return errors.New("sorted set zrange:" + err.Error())
	}
	jsonByte, err := json.Marshal(valueArr)
	fmt.Println(string(jsonByte))
	err = json.Unmarshal(jsonByte, &dest)
	return nil
}

//Count
func (store *SortedSet) Count(args ...interface{}) (count int64, err error) {
	if len(args) == 0 {
		return 0, errors.New("args is nil")
	}
	key, b := args[0].(string)
	if !b {
		return 0, errors.New("one arg is nil")
	}
	c, err := conn.GetRedisClient().ZCard(key)
	if err != nil {
		return 0, errors.New("sorted set zcard:" + err.Error())
	}
	return int64(c), nil
}
