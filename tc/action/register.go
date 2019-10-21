package action

import (
	"context"
	"goseata/proto"
	"goseata/tc/consts"
	error2 "goseata/tc/error"
	"goseata/tc/lock"
	"goseata/tc/log"
	"goseata/tc/mysql"
	"strconv"

	"github.com/orzzzli/orzconfiger"
)

type TcServer struct{}

func (s *TcServer) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterReply, error) {
	//打印trace
	log.New(req.TraceId, "register", req.RequestPath).ToLog()

	firstService := req.RequestPath.Infos[0]

	appid, find := orzconfiger.GetString("service", "appid")
	if !find {
		return handleRegError(req.TraceId, consts.ConfigError, nil, "config not found service appid")
	}
	lockManager := lock.New(appid, req.Tid, nil)

	//检查锁是否存在
	for _, v := range req.Locks {
		//绑定锁
		lockManager.SetPLock(v)
		//锁结构：连接标记.数据库.表.主键.主键值
		success, err, err2 := lockManager.Lock()
		if err != nil {
			return handleRegError(req.TraceId, consts.LockError, err, "try to lock error.", appid, req.Tid, v)
		}
		if err2 != nil {
			return handleRegError(req.TraceId, consts.LockError, err2, "clear lock error.", appid, req.Tid, v)
		}
		//设置锁失败，锁已存在
		if !success {
			return handleRegError(req.TraceId, consts.SetLockFail, nil, "set lock fail. lock exist.", lockManager.ToStr(), v)
		}
	}

	log.New(req.TraceId, "register", "set all lock success.", req.Locks, lockManager.ToStr()).ToLog()

	transactionDB, find := mysql.DBPoolsInstance.GetPool("db-transaction")
	if !find {
		err := lockManager.ClearLocks()
		if err != nil {
			return handleRegError(req.TraceId, consts.LockError, err, "clear lock error.", lockManager.ToStr())
		}
		return handleRegError(req.TraceId, consts.DBError, nil, "transaction db not found.")
	}

	//生成分支事务
	result, err := transactionDB.Exec("INSERT INTO `local_transaction`(tid,appid,name)VALUES (?,?,?)", req.Tid, firstService.Appid, firstService.Name)
	if err != nil {
		err2 := lockManager.ClearLocks()
		if err2 != nil {
			return handleRegError(req.TraceId, consts.LockError, err2, "clear lock error.", lockManager.ToStr())
		}
		return handleRegError(req.TraceId, consts.DBError, err, "insert local transaction error.")
	}
	ltid, _ := result.LastInsertId()
	ltidStr := strconv.FormatInt(ltid, 10)

	log.New(req.TraceId, "register", "insert local transaction success.", req.Tid, firstService.Appid, firstService.Name, ltid).ToLog()

	return &proto.RegisterReply{
		ReplyInfo: &proto.ReplyInfo{
			Code:    0,
			Message: "success",
		},
		Lock:    true,
		Ltid:    ltidStr,
		TraceId: req.TraceId,
	}, nil
}

func handleRegError(traceId string, code int, err error, extends ...interface{}) (*proto.RegisterReply, error) {
	errorObj := error2.New(traceId, code, err, extends)
	errorObj.ToLog()
	return &proto.RegisterReply{
		ReplyInfo: errorObj.ToReplyInfo(),
		Lock:      false,
		TraceId:   traceId,
	}, nil
}
