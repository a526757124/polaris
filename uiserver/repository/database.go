package repository

import "github.com/a526757124/polaris/uiserver/repository/redis"

type Database interface {
	Insert(args ...interface{}) (n int64, err error)
	Update(args ...interface{}) (n int64, err error)
	Delete(args ...interface{}) (n int64, err error)
	Find(dest interface{}, args ...interface{}) error
	FindOne(dest interface{}, args ...interface{}) error
	FindByPage(dest interface{}, skip, limit int64, args ...interface{}) error
	Count(args ...interface{}) (count int64, err error)
}

func GetDataBase() Database {
	return &redis.SortedSet{}
}
