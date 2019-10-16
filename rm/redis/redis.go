package redis

import (
	"errors"
	"github.com/gomodule/redigo/redis"
	"goseata/util"
	"log"
	"time"
)

var GlobalRedisPool *redis.Pool

func init() {
	NewRedisPool("redis://127.0.0.1:6379","",10,300)
	log.Println("redis worked")
}

// NewRedisPool初始化连接池
func NewRedisPool(url string, password string, idle int, idleTime int) {
	GlobalRedisPool = &redis.Pool{
		MaxIdle:     idle,
		IdleTimeout: time.Duration(idleTime) * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialURL(url)
			if err != nil {
				panic(err)
			}
			if password != "" {
				//验证redis密码
				if _, authErr := c.Do("AUTH", password); authErr != nil {
					panic(err)
				}
			}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			if err != nil {
				panic(err)
			}
			return nil
		},
	}
}

//expire
func Expire(k string, ex int) error {
	if GlobalRedisPool == nil {
		return errors.New("redis pool is not init.")
	}
	conn := GlobalRedisPool.Get()
	defer conn.Close()
	var err error
	_, err = conn.Do("EXPIRE", k, ex)
	return err
}

//string
func Set(k string, v string, ex int) error {
	if GlobalRedisPool == nil {
		return errors.New("redis pool is not init.")
	}
	conn := GlobalRedisPool.Get()
	defer conn.Close()
	var err error
	if ex <= 0 {
		_, err = conn.Do("SET", k, v)
	} else {
		_, err = conn.Do("SET", k, v, "EX", ex)
	}
	return err
}
func SetNx(k string, v string, ex int) (bool,error) {
	if GlobalRedisPool == nil {
		return false,errors.New("redis pool is not init.")
	}
	conn := GlobalRedisPool.Get()
	defer conn.Close()
	var err error
	var res interface{}
	if ex <= 0 {
		res, err = conn.Do("SETNX", k, v)
	} else {
		res, err = conn.Do("SETNX", k, v, "EX", ex)
	}
	if err != nil {
		return false, err
	}
	if res.(int64) == 1 {
		return true,err
	}else{
		return false,err
	}
}
func Get(k string) (string, bool, error) {
	if GlobalRedisPool == nil {
		return "", false, errors.New("redis pool is not init.")
	}
	conn := GlobalRedisPool.Get()
	defer conn.Close()
	res, err := conn.Do("GET", k)
	resOp := ""
	find := false
	if res != nil {
		resOp = string(res.([]uint8))
		find = true
	}
	return resOp, find, err
}

//list
func LPush(k string, v string) error {
	if GlobalRedisPool == nil {
		return errors.New("redis pool is not init.")
	}
	conn := GlobalRedisPool.Get()
	defer conn.Close()
	var err error
	_, err = conn.Do("LPUSH", k, v)
	return err
}
func RPop(k string) (string, bool, error) {
	if GlobalRedisPool == nil {
		return "", false, errors.New("redis pool is not init.")
	}
	conn := GlobalRedisPool.Get()
	defer conn.Close()
	res, err := conn.Do("RPOP", k)
	resOp := ""
	find := false
	if res != nil {
		resOp = string(res.([]uint8))
		find = true
	}
	return resOp, find, err
}
func LRange(k string, start int, end int) ([]string, error) {
	if GlobalRedisPool == nil {
		return nil, errors.New("redis pool is not init.")
	}
	conn := GlobalRedisPool.Get()
	defer conn.Close()
	res, err := conn.Do("LRANGE", k, start, end)
	var resOp []string
	for _,v := range res.([]interface{}) {
		resOp = append(resOp,string(v.([]uint8)))
	}
	return resOp, err
}

//SortedSet
func ZAdd(key string, k string, v float32) error {
	if GlobalRedisPool == nil {
		return errors.New("redis pool is not init.")
	}
	conn := GlobalRedisPool.Get()
	defer conn.Close()
	var err error
	_, err = conn.Do("ZADD", key, v, k)
	return err
}
func ZRevRank(key string, k string) (int, bool, error) {
	if GlobalRedisPool == nil {
		return 0, false, errors.New("redis pool is not init.")
	}
	conn := GlobalRedisPool.Get()
	defer conn.Close()
	resultOp := 0
	find := false
	result, err := conn.Do("ZREVRANK", key, k)
	if result != nil {
		resultOp, _ = util.Int64to32(result.(int64))
		find = true
	}
	return resultOp, find, err
}

//del
func Del(k string) error {
	if GlobalRedisPool == nil {
		return errors.New("redis pool is not init.")
	}
	conn := GlobalRedisPool.Get()
	defer conn.Close()
	var err error
	_, err = conn.Do("DEL", k)
	return err
}