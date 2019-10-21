package lock

import (
	"goseata/proto"
	"goseata/tc/redis"
	"testing"
)

const (
	RedisURL      = "redis://127.0.0.1:6379"
	RedisPass     = ""
	MaxIdleNumber = 1
	MaxIdleTime   = 60
)

func TestNew(t *testing.T) {
	redis.NewRedisPool(RedisURL, RedisPass, MaxIdleNumber, MaxIdleTime)

	lock := New("aaa", "bbb", nil)
	println(lock.ToStr())
	lock.getTempLockListKey()
	lock.pLock = &proto.Lock{
		Connection: "a",
		Database:   "b",
		Table:      "c",
		PrimaryK:   "d",
		PrimaryV:   "e",
	}
	println(lock.ToStr())
	success, err, err2 := lock.Lock()
	println(success, err, err2)
	err = lock.ClearLocks()
	println(err)
}
