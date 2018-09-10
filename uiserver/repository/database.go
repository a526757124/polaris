package repository

type Databse interface {
	Insert(args ...interface{}) (n int64, err error)
	Update(args ...interface{}) (n int64, err error)
	Delete(args ...interface{}) (n int64, err error)
	Find(dest interface{}, args ...interface{}) error
	FindOne(dest interface{}, args ...interface{}) error
	FindByPage(dest interface{}, skip, limit int, args ...interface{}) error
	Count(args ...interface{}) (count int64, err error)
}
