package action

import (
	"context"
	"goseata/proto"
	"goseata/tc/consts"
	error2 "goseata/tc/error"
	"goseata/tc/lock"
	"goseata/tc/log"
	"goseata/tc/model"
	"goseata/tc/mysql"

	"github.com/orzzzli/orzconfiger"
)

func (s *TcServer) GlobalCommit(ctx context.Context, req *proto.GlobalCommitRequest) (*proto.GlobalCommitReply, error) {
	//打印trace
	log.New(req.TraceId, "globalCommit", req.RequestPath).ToLog()

	transactionDB, find := mysql.DBPoolsInstance.GetPool("db-transaction")
	if !find {
		return handleGcError(req.TraceId, consts.DBError, nil, "transaction db not found.")
	}

	//检查该事务下的分支事务是否全部完成
	var localTransactions []*model.LocalTransaction
	err := transactionDB.Select(&localTransactions, "SELECT * FROM `local_transaction` where tid = ? and status <> ?", req.Tid, model.LocalTransactionStatusCommitted)
	if err != nil {
		return handleGcError(req.TraceId, consts.DBError, err, "select local transaction error.", req.Tid, model.LocalTransactionStatusCommitted)
	}
	if len(localTransactions) != 0 {
		return handleGcError(req.TraceId, consts.DBError, nil, "local transactions not all done.", req.Tid)
	}

	//清除锁
	appid, find := orzconfiger.GetString("service", "appid")
	if !find {
		return handleGcError(req.TraceId, consts.ConfigError, nil, "appid not found.")
	}
	lockManager := lock.New(appid, req.Tid, nil)
	err = lockManager.ClearLocks()
	if err != nil {
		return handleGcError(req.TraceId, consts.LockError, err, "clear lock error.", lockManager.ToStr())
	}

	return &proto.GlobalCommitReply{
		ReplyInfo: &proto.ReplyInfo{
			Code:    0,
			Message: "success",
		},
		TraceId: req.TraceId,
	}, nil
}

func handleGcError(traceId string, code int, err error, extends ...interface{}) (*proto.GlobalCommitReply, error) {
	errorObj := error2.New(traceId, code, err, extends)
	errorObj.ToLog()
	return &proto.GlobalCommitReply{
		ReplyInfo: errorObj.ToReplyInfo(),
		TraceId:   traceId,
	}, nil
}
