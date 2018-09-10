package redis

//sorted set 存储
type SortedSet struct {
}

//Insert
func (store *SortedSet) Insert(args ...interface{}) (n int64, err error) {
	//conn.GetRedisClient().ZAdd()

	return 0, nil
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
func (store *SortedSet) FindByPage(dest interface{}, skip, limit int, args ...interface{}) error {
	return nil
}

//Count
func (store *SortedSet) Count(args ...interface{}) (count int64, err error) {
	return 0, nil
}
