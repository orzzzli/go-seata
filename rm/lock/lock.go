package lock

import "goseata/rm/redis"

func getLocalLockKey(appid string, tid string) string {
	return appid + "." + tid
}

func getLockStr(connect string, database string, table string, primaryK string, primaryV string) string {
	return connect + "|" + database + "|" + table + "|" + primaryK + "|" + primaryV
}

func SetLocalLock(tid string, connect string, database string, table string, primaryK string, primaryV string) error {
	localKey := getLocalLockKey("100", tid)
	lockStr := getLockStr(connect, database, table, primaryK, primaryV)
	err := redis.LPush(localKey, lockStr)
	return err
}

func GetLocalLocks(tid string) ([]string, error) {
	localKey := getLocalLockKey("100", tid)
	locks, err := redis.LRange(localKey, 0, -1)
	if err != nil {
		return nil, err
	}
	return locks, nil
}

func RmLocalLock(tid string) error {
	localKey := getLocalLockKey("100", tid)
	//清除本地锁信息
	err := redis.Del(localKey)
	return err
}
