package lock

import (
	"goseata/proto"
	"goseata/tc/redis"
)

type Lock struct {
	appid       string
	tid         string
	pLock       *proto.Lock
	TempListKey string
}

func New(appid string, tid string, plock *proto.Lock) *Lock {
	lock := &Lock{
		appid: appid,
		tid:   tid,
		pLock: plock,
	}
	if lock.pLock == nil {
		lock.pLock = &proto.Lock{}
	}
	lock.getTempLockListKey()
	return lock
}

/*
	pLock setter.
*/
func (l *Lock) SetPLock(plock *proto.Lock) {
	l.pLock = plock
}

/*
	Lock str, row lock.
	Format: connection|database|table|primaryKey|primaryValue
*/
func (l *Lock) ToStr() string {
	return l.pLock.Connection + "|" + l.pLock.Database + "|" + l.pLock.Table + "|" + l.pLock.PrimaryK + "|" + l.pLock.PrimaryV
}

/*
	Create tc local temp lock list key.
	Temp lock list use to storage all this tid contain locks.
	This list storage in redis.

	Format: tc.appid.tid
*/
func (l *Lock) getTempLockListKey() {
	l.TempListKey = "tc." + l.appid + "." + l.tid
}

/*
	LPush lock str to local temp lock list.
*/
func (l *Lock) setTempLock() error {
	err := redis.LPush(l.TempListKey, l.ToStr())
	return err
}

/*
	Actual lock action func.
	Use redis setNx to simulate global lock.
*/
func (l *Lock) Lock() (lockSuccess bool, setLockErr error, clearLockErr error) {
	success, err := redis.SetNx(l.ToStr(), "1", 0)
	//set lock error. clear this session all lock.
	if err != nil {
		//clear already set lock avoid dead lock.
		err2 := l.ClearLocks()
		return false, err, err2
	}
	//set lock fail. clear this session all lock.
	if !success {
		err2 := l.ClearLocks()
		return false, nil, err2
	}
	if success {
		err = l.setTempLock()
		if err != nil {
			err2 := l.ClearLocks()
			return false, err, err2
		}
	}
	return success, err, nil
}

/*
	Clear all type locks.
	Contain db row locks and temp local lock list.

	Todo:This error must take care, it may cause dead lock.
*/
func (l *Lock) ClearLocks() error {
	locks, err := redis.LRange(l.TempListKey, 0, -1)
	if err != nil {
		return err
	}
	for _, v := range locks {
		err := redis.Del(v)
		if err != nil {
			return err
		}
	}
	//clear temp local lock list.
	err = redis.Del(l.TempListKey)
	if err != nil {
		return err
	}
	return nil
}
