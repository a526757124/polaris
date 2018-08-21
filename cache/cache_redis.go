package cache

import (
	"strconv"
	"strings"
	"github.com/devfeel/polaris/util/redisx"
	"github.com/devfeel/hystrix"
)

const(
	HystrixErrorCount = 50
)
// RedisCache is redis cache adapter.
// it contains serverIp for redis conn.
type redisCache struct {

	hystrix hystrix.Hystrix

	serverUrl string //connection string, like "redis://:password@10.0.1.11:6379/0"
	// Maximum number of idle connections in the pool.
	maxIdle int
	// Maximum number of connections allocated by the pool at a given time.
	// When zero, there is no limit on the number of connections in the pool.
	maxActive int

	//use to backup server
	backupServerUrl string
	backupMaxIdle int
	backupMaxActive int
}

// NewRedisCache returns a new *RedisCache.
func NewRedisCache(serverUrl string, maxIdle int, maxActive int) *redisCache {
	cache := redisCache{serverUrl: serverUrl, maxIdle:maxIdle, maxActive:maxActive}
	cache.hystrix = hystrix.NewHystrix(cache.checkRedisAlive, nil)
	cache.hystrix.SetMaxFailedNumber(HystrixErrorCount)
	cache.hystrix.Do()
	return &cache
}

// SetBackupServer set backup redis server, only use to read
func (ca *redisCache) SetBackupServer(serverUrl string, maxIdle int, maxActive int){
	ca.backupServerUrl = serverUrl
	ca.backupMaxActive = maxActive
	ca.backupMaxIdle = maxIdle
}

// Exists check item exist in redis cache.
func (ca *redisCache) Exists(key string) (bool, error) {
	client := ca.getRedisClient()
	exists, err := client.Exists(key)
	if ca.checkConnErrorAndNeedRetry(err){
		client = ca.getBackupRedis()
		return client.Exists(key)
	}
	return exists, err
}

// Incr increase int64 counter in redis cache.
func (ca *redisCache) Incr(key string) (int64, error) {
	client := ca.getDefaultRedis()
	val, err := client.INCR(key)
	if ca.checkConnErrorAndNeedRetry(err){
		client = ca.getBackupRedis()
		val, err = client.INCR(key)
	}
	if err != nil {
		return 0, err
	}
	return int64(val), nil
}

// Decr decrease counter in redis cache.
func (ca *redisCache) Decr(key string) (int64, error) {
	client := ca.getDefaultRedis()
	val, err := client.DECR(key)
	if ca.checkConnErrorAndNeedRetry(err){
		client = ca.getBackupRedis()
		val, err = client.DECR(key)
	}
	if err != nil {
		return 0, err
	}
	return int64(val), nil
}

// Get cache from redis cache.
// if non-existed or expired, return nil.
func (ca *redisCache) Get(key string) (interface{}, error) {
	client := ca.getRedisClient()
	reply, err := client.GetObj(key)
	if ca.checkConnErrorAndNeedRetry(err){
		client = ca.getBackupRedis()
		return client.GetObj(key)
	}
	return reply, err
}

// GetString returns value string format by given key
// if non-existed or expired, return "".
func (ca *redisCache) GetString(key string) (string, error) {
	client := ca.getRedisClient()
	reply, err := client.Get(key)
	if ca.checkConnErrorAndNeedRetry(err){
		client = ca.getBackupRedis()
		return client.Get(key)
	}
	return reply, err
}

// GetInt returns value int format by given key
// if non-existed or expired, return nil.
func (ca *redisCache) GetInt(key string) (int, error) {
	v, err := ca.GetString(key)
	if err != nil || v == "" {
		return 0, err
	} else {
		i, e := strconv.Atoi(v)
		if e != nil {
			return 0, err
		} else {
			return i, nil
		}
	}
}

// GetInt64 returns value int64 format by given key
// if non-existed or expired, return nil.
func (ca *redisCache) GetInt64(key string) (int64, error) {
	v, err := ca.GetString(key)
	if err != nil || v == "" {
		return ZeroInt64, err
	} else {
		i, e := strconv.ParseInt(v, 10, 64)
		if e != nil {
			return ZeroInt64, err
		} else {
			return i, nil
		}
	}
}

// Set cache to redis.
// ttl is second, if ttl is 0, it will be forever.
func (ca *redisCache) Set(key string, value interface{}, ttl int64) error {
	var err error
	client := redisx.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	if ttl > 0{
		_, err = client.SetWithExpire(key, value, ttl)
		if ca.checkConnErrorAndNeedRetry(err){
			client = ca.getBackupRedis()
			_, err = client.SetWithExpire(key, value, ttl)
		}
	}else{
		_, err = client.Set(key, value)
		if ca.checkConnErrorAndNeedRetry(err){
			client = ca.getBackupRedis()
			_, err = client.Set(key, value)
		}
	}
	return err
}

// Delete item in redis cacha.
// if not exists, we think it's success
func (ca *redisCache) Delete(key string) error {
	client := redisx.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	_, err := client.Del(key)
	if ca.checkConnErrorAndNeedRetry(err){
		client = ca.getBackupRedis()
		_, err = client.Del(key)
	}
	return err
}

// Expire Set a timeout on key. After the timeout has expired, the key will automatically be deleted.
func (ca *redisCache) Expire(key string, timeOutSeconds int) (int, error){
	client := redisx.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	reply, err := client.Expire(key, timeOutSeconds)
	if ca.checkConnErrorAndNeedRetry(err){
		client = ca.getBackupRedis()
		reply, err = client.Expire(key, timeOutSeconds)
	}
	return reply, err
}

// GetJsonObj get obj with SetJsonObj key
func (ca *redisCache) GetJsonObj(key string, result interface{})error {
	client := ca.getRedisClient()
	err := client.GetJsonObj(key, result)
	if ca.checkConnErrorAndNeedRetry(err){
		client = ca.getBackupRedis()
		return client.GetJsonObj(key, result)
	}
	return err
}

// SetJsonObj set obj use json encode string
func (ca *redisCache) SetJsonObj(key string, val interface{}) (interface{}, error){
	client := redisx.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	reply, err := client.SetJsonObj(key, val)
	if ca.checkConnErrorAndNeedRetry(err){
		client = ca.getBackupRedis()
		reply, err = client.SetJsonObj(key, val)
	}
	return reply, err
}

/*---------- Hash -----------*/
// HGet Returns the value associated with field in the hash stored at key.
func (ca *redisCache) HGet(key, field string) (string, error) {
	client := ca.getRedisClient()
	reply, err := client.HGet(key, field)
	if ca.checkConnErrorAndNeedRetry(err){
		client = ca.getBackupRedis()
		return client.HGet(key, field)
	}
	return reply, err
}


// HMGet Returns the values associated with the specified fields in the hash stored at key.
func (ca *redisCache) HMGet(hashID string, field ...interface{}) ([]string, error) {
	client := ca.getRedisClient()
	reply, err := client.HMGet(hashID, field...)
	if ca.checkConnErrorAndNeedRetry(err){
		client = ca.getBackupRedis()
		return client.HMGet(hashID, field...)
	}
	return reply, err
}

// HGetAll Returns all fields and values of the hash stored at key
func (ca *redisCache) HGetAll(key string) (map[string]string, error) {
	client := ca.getRedisClient()
	reply, err := client.HGetAll(key)
	if ca.checkConnErrorAndNeedRetry(err){
		client = ca.getBackupRedis()
		return client.HGetAll(key)
	}
	return reply, err
}

// HSet Sets field in the hash stored at key to value. If key does not exist, a new key holding a hash is created.
// If field already exists in the hash, it is overwritten.
func (ca *redisCache) HSet(key, field, value string) error {
	client := redisx.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	err := client.HSet(key, field, value)
	if ca.checkConnErrorAndNeedRetry(err){
		client = ca.getBackupRedis()
		err = client.HSet(key, field, value)
	}
	return err
}
// HDel Removes the specified fields from the hash stored at key.
func (ca *redisCache) HDel(key string, field ...interface{}) (int, error){
	client := redisx.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	reply, err := client.HDel(key, field...)
	if ca.checkConnErrorAndNeedRetry(err){
		client = ca.getBackupRedis()
		reply, err = client.HDel(key, field...)
	}
	return reply, err
}
// HExists Returns if field is an existing field in the hash stored at key
func (ca *redisCache) HExists (key string, field string) (int, error){
	client := ca.getRedisClient()
	reply, err := client.HExist(key, field)
	if ca.checkConnErrorAndNeedRetry(err){
		client = ca.getBackupRedis()
		return client.HExist(key, field)
	}
	return reply, err
}
// HSetNX Sets field in the hash stored at key to value, only if field does not yet exist
func (ca *redisCache) HSetNX(key string, field string, value string) (string, error){
	client := redisx.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	reply, err :=  client.HSetNX(key, field, value)
	if ca.checkConnErrorAndNeedRetry(err){
		client = ca.getBackupRedis()
		reply, err =  client.HSetNX(key, field, value)
	}
	return reply, err
}
// HIncrBy Increments the number stored at field in the hash stored at key by increment.
func (ca *redisCache) HIncrBy(key string, field string, increment int) (int, error){
	client := redisx.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	reply, err :=  client.HIncrBy(key, field, increment)
	if ca.checkConnErrorAndNeedRetry(err){
		client = ca.getBackupRedis()
		reply, err =  client.HIncrBy(key, field, increment)
	}
	return reply, err
}

// HKeys Returns all field names in the hash stored at key.
func (ca *redisCache) HKeys(key string) ([]string, error){
	client := ca.getRedisClient()
	reply, err := client.HKeys(key)
	if ca.checkConnErrorAndNeedRetry(err){
		client = ca.getBackupRedis()
		return client.HKeys(key)
	}
	return reply, err
}
// HLen Returns the number of fields contained in the hash stored at key
func (ca *redisCache) HLen(key string) (int, error){
	client := ca.getRedisClient()
	reply, err := client.HLen(key)
	if ca.checkConnErrorAndNeedRetry(err){
		client = ca.getBackupRedis()
		return client.HLen(key)
	}
	return reply, err
}


/*---------- List -----------*/

// BRPOP is a blocking list pop primitive
// It is the blocking version of RPOP because it blocks the connection when there are no elements to pop from any of the given lists
func (ca *redisCache) BRPop(key ...interface{}) (map[string]string, error){
	client := redisx.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	reply, err := client.BRPop(key)
	if ca.checkConnErrorAndNeedRetry(err){
		client = ca.getBackupRedis()
		reply, err = client.BRPop(key)
	}
	return reply, err
}
// LLen return length of list
func (ca *redisCache) LLen(key string) (int, error){
	client := redisx.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	reply, err :=  client.LLen(key)
	if ca.checkConnErrorAndNeedRetry(err){
		client = ca.getBackupRedis()
		reply, err = client.LLen(key)
	}
	return reply, err
}

// LPush Insert all the specified values at the head of the list stored at key
func (ca *redisCache) LPush(key string, value ...interface{}) (int, error){
	client := redisx.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	reply, err :=  client.LPush(key, value)
	if ca.checkConnErrorAndNeedRetry(err){
		client = ca.getBackupRedis()
		reply, err = client.LPush(key, value)
	}
	return reply, err
}


//****************** 全局操作 ***********************
// Ping ping command, if success return pong
func  (ca *redisCache) Ping()(string,error){
	client := redisx.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
	return client.Ping()
}




// getReadRedisClient get read mode redis client
func (ca *redisCache) getRedisClient() *redisx.RedisClient{
	if ca.hystrix.IsHystrix(){
		if ca.backupServerUrl != "" {
			return ca.getBackupRedis()
		}
	}
	return ca.getDefaultRedis()
}

// getRedisClient get default redis client
func (ca *redisCache) getDefaultRedis() *redisx.RedisClient{
	return redisx.GetRedisClient(ca.serverUrl, ca.maxIdle, ca.maxActive)
}

func (ca *redisCache) getBackupRedis() *redisx.RedisClient{
	return redisx.GetRedisClient(ca.backupServerUrl, ca.backupMaxIdle, ca.backupMaxActive)
}


// checkConnErrorAndNeedRetry check err is Conn error and is need to retry
// if current client is hystrix, no need retry, because in getReadRedisClient already use backUp redis
func (ca *redisCache) checkConnErrorAndNeedRetry(err error) bool{
	if err == nil{
		return false
	}
	if strings.Index(err.Error(), "no such host") >= 0 ||
		strings.Index(err.Error(), "No connection could be made because the target machine actively refused it") >= 0 ||
		strings.Index(err.Error(), "A connection attempt failed because the connected party did not properly respond after a period of time") >= 0 {
		ca.hystrix.GetCounter().Inc(1)
		//if is hystrix, not to retry, because in getReadRedisClient already use backUp redis
		if ca.hystrix.IsHystrix(){
			return false
		}
		if ca.backupServerUrl == ""{
			return false
		}
		return true
	}
	return false
}

// checkRedisAlive check redis is alive use ping
// if set readonly redis, check readonly redis
// if not set readonly redis, check default redis
func (ca *redisCache) checkRedisAlive() bool{
	isAlive := false
	redisClient := ca.getDefaultRedis()
	for i := 0;i<=5;i++ {
		reply, err := redisClient.Ping()
		//fmt.Println(time.Now(), "checkAliveDefaultRedis Ping", reply, err)
		if err != nil {
			isAlive = false
			break
		}
		if reply != "PONG" {
			isAlive = false
			break
		}
		isAlive = true
		continue
	}
	return isAlive
}

