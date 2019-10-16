package action

import (
	"context"
	"errors"
	"goseata/proto"
	"goseata/tc/lock"
	"goseata/tc/model"
	"goseata/tc/mysql"
	"goseata/util"
)

func (s *TcServer) GlobalCommit(ctx context.Context, req *proto.GlobalCommitRequest) (*proto.GlobalCommitReply, error) {
	//打印trace
	util.Trace(req.TraceId, "globalCommit", req.RequestPath)

	transactionDB, find := mysql.DBPoolsInstance.GetPool("db-transaction")
	if !find {
		err := errors.New("transaction db not found")
		return handleGcError(2000, req.TraceId, err, ""), nil
	}

	//检查该事务下的分支事务是否全部完成
	var localTransactions []*model.LocalTransaction
	err := transactionDB.Select(&localTransactions, "SELECT * FROM `local_transaction` where tid = ? and status <> ?", req.Tid, model.LocalTransactionStatusCommitted)
	if err != nil {
		return handleGcError(2100, req.TraceId, err, "select local transaction error. tid is"+req.Tid), nil
	}
	if len(localTransactions) != 0 {
		err = errors.New("current transaction " + req.Tid + " local transactions not all done")
		return handleGcError(2200, req.TraceId, err, ""), nil
	}

	//清除锁
	err = lock.RmLocks(req.Tid)
	if err != nil {
		return handleGcError(2300, req.TraceId, err, "clear lock error."), nil
	}

	return &proto.GlobalCommitReply{
		ReplyInfo: reply(0, "success"),
		TraceId:   req.TraceId,
	}, nil
}

func handleGcError(code int, traceId string, err error, info string) *proto.GlobalCommitReply {
	if info != "" {
		err = errors.New(err.Error() + " Info:" + info)
	}
	util.LogError(traceId, err)
	return &proto.GlobalCommitReply{
		ReplyInfo: reply(code, err.Error()),
		TraceId:   traceId,
	}
}
