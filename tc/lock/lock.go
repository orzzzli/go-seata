package lock

import (
	"errors"
	"goseata/tc/redis"

	"github.com/orzzzli/orzconfiger"
)

func getTempLockKey(appid string, tid string) string {
	return "tc." + appid + "." + tid
}

func setTempLock(key string, lockStr string) error {
	err := redis.LPush(key, lockStr)
	return err
}

func getLockStr(connect string, database string, table string, primaryK string, primaryV string) string {
	return connect + "|" + database + "|" + table + "|" + primaryK + "|" + primaryV
}

func SetLock(tid string, connect string, database string, table string, primaryK string, primaryV string) (bool, string, error) {
	lockStr := getLockStr(connect, database, table, primaryK, primaryV)
	success, err := redis.SetNx(lockStr, "1", 0)
	if success {
		appid, find := orzconfiger.GetString("service", "appid")
		if !find {
			err = errors.New("read config error: cant get appid from config")
			//避免死锁
			redis.Del(lockStr)
			return false, "", err
		}
		tempLockKey := getTempLockKey(appid, tid)
		err = setTempLock(tempLockKey, lockStr)
		if err != nil {
			//避免死锁
			redis.Del(lockStr)
			return false, "", err
		}
	}
	return success, lockStr, err
}

//todo:出错重点关注，清除锁失败会引起死锁
func RmLocks(tid string) error {
	appid, find := orzconfiger.GetString("service", "appid")
	if !find {
		err := errors.New("read config error: cant get appid from config")
		return err
	}
	lockKey := getTempLockKey(appid, tid)
	locks, err := redis.LRange(lockKey, 0, -1)
	if err != nil {
		return err
	}
	for _, v := range locks {
		err := redis.Del(v)
		if err != nil {
			return err
		}
	}
	//清除本地锁信息
	err = redis.Del(lockKey)
	if err != nil {
		return err
	}
	return nil
}
