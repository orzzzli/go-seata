package action

import (
	"context"
	"errors"
	"goseata/proto"
	"goseata/tc/lock"
	"goseata/tc/mysql"
	"goseata/util"
	"strconv"
)

type TcServer struct{}

func (s *TcServer) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterReply, error) {
	//打印trace
	util.Trace(req.TraceId, "register", req.RequestPath)

	firstService := req.RequestPath.Infos[0]

	//检查锁是否存在
	for _, v := range req.Locks {
		//锁结构：连接标记.数据库.表.主键.主键值
		success, lockStr, err := lock.SetLock(req.Tid, v.Connection, v.Database, v.Table, v.PrimaryK, v.PrimaryV)
		if err != nil {
			err = lock.RmLocks(req.Tid)
			if err != nil {
				return handleRegError(101, req.TraceId, err, "clear lock error!!!!"), nil
			}
			return handleRegError(100, req.TraceId, err, "lock is"+lockStr), nil
		}
		//设置锁失败，锁已存在
		if !success {
			err = lock.RmLocks(req.Tid)
			if err != nil {
				return handleRegError(101, req.TraceId, err, "clear lock error!!!!"), nil
			}
			err = errors.New("set lock fail. lock exist " + lockStr)
			//这种情况非错误，因此code为0，让子服务触发重试
			return handleRegError(0, req.TraceId, err, ""), nil
		}
	}

	transactionDB, find := mysql.DBPoolsInstance.GetPool("db-transaction")
	if !find {
		err := errors.New("transaction db not found")
		err = lock.RmLocks(req.Tid)
		if err != nil {
			return handleRegError(101, req.TraceId, err, "clear lock error!!!!"), nil
		}
		return handleRegError(300, req.TraceId, err, ""), nil
	}

	//生成分支事务
	result, err := transactionDB.Exec("INSERT INTO `local_transaction`(tid,appid,name)VALUES (?,?,?)", req.Tid, firstService.Appid, firstService.Name)
	if err != nil {
		err = lock.RmLocks(req.Tid)
		if err != nil {
			return handleRegError(101, req.TraceId, err, "clear lock error!!!!"), nil
		}
		return handleRegError(400, req.TraceId, err, "insert local transaction error"), nil
	}
	ltid, _ := result.LastInsertId()
	ltidStr := strconv.FormatInt(ltid, 10)

	return &proto.RegisterReply{
		ReplyInfo: reply(0, "success"),
		Lock:      true,
		Ltid:      ltidStr,
		TraceId:   req.TraceId,
	}, nil
}

func handleRegError(code int, traceId string, err error, info string) *proto.RegisterReply {
	if info != "" {
		err = errors.New(err.Error() + " Info:" + info)
	}
	util.LogError(traceId, err)
	return &proto.RegisterReply{
		ReplyInfo: reply(code, err.Error()),
		Lock:      false,
		TraceId:   traceId,
	}
}

func reply(code int, msg string) *proto.ReplyInfo {
	return &proto.ReplyInfo{
		Code:    int32(code),
		Message: msg,
	}
}
