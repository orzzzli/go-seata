package action

import (
	"context"
	"encoding/json"
	"errors"
	"goseata/proto"
	"goseata/tc/lock"
	"goseata/tc/model"
	"goseata/tc/mysql"
	"goseata/util"
	"strings"

	"github.com/orzzzli/orzconfiger"
)

func (s *TcServer) GlobalRollback(ctx context.Context, req *proto.GlobalRollbackRequest) (*proto.GlobalRollbackReply, error) {
	//打印trace
	util.Trace(req.TraceId, "globalRollback", req.RequestPath)

	transactionDB, find := mysql.DBPoolsInstance.GetPool("db-transaction")
	if !find {
		err := errors.New("transaction db not found")
		return handleGrError(3000, req.TraceId, err, ""), nil
	}

	var localTransactions []*model.LocalTransaction
	err := transactionDB.Select(&localTransactions, "SELECT * FROM `local_transaction` where tid = ?", req.Tid)
	if err != nil {
		return handleGrError(3100, req.TraceId, err, "select local transaction error. tid is"+req.Tid), nil
	}

	for _, v := range localTransactions {
		appid := v.Appid
		connect, find := orzconfiger.GetString(appid, "connect")
		if !find {
			err = errors.New("get connect string error.appid is " + appid + " tid is" + req.Tid)
			return handleGrError(3200, req.TraceId, err, ""), nil
		}
		err := rollback(req.TraceId, connect, req.Tid)
		if err != nil {
			return handleGrError(3300, req.TraceId, err, "rollback error.connect is "+connect+" tid is"+req.Tid), nil
		}
		//更新分支事务状态
		_, err = transactionDB.Exec("update `local_transaction` set status = ? where id = ?", model.LocalTransactionStatusRollbacked, v.Id)
		if err != nil {
			return handleGrError(3400, req.TraceId, err, "update local transaction status error"), nil
		}
	}

	//清除锁
	err = lock.RmLocks(req.Tid)
	if err != nil {
		return handleGrError(3500, req.TraceId, err, "clear lock error."), nil
	}

	return &proto.GlobalRollbackReply{
		ReplyInfo: reply(0, "success"),
		TraceId:   req.TraceId,
	}, nil
}

func rollback(traceId string, connectionStr string, tid string) error {
	oneDB, find := mysql.DBPoolsInstance.GetPool(connectionStr)
	if !find {
		err := errors.New(connectionStr + " db not found")
		return err
	}
	var logs []*model.TransactionLog
	err := oneDB.Select(&logs, "SELECT * FROM `transcation_log` where tid = ?", tid)
	if err != nil {
		return err
	}
	for _, v := range logs {
		sqlType := v.Type
		table := v.Table
		primaryK := v.PrimaryKey
		primaryV := v.PrimaryValue
		beforeJson := v.BeforeCol
		tempMap := make(map[string]string)
		err := json.Unmarshal([]byte(beforeJson), &tempMap)
		if err != nil {
			return err
		}
		undoSql := ""
		if sqlType == model.TransactionLogTypeInsert {
			undoSql = "DELETE FROM `" + table + "` WHERE " + primaryK + " = '" + primaryV + "' LIMIT 1;"
		} else if sqlType == model.TransactionLogTypeUpdate {
			undoSql = "UPDATE `" + table + "` SET "
			for key, value := range tempMap {
				undoSql += key + "=" + value + ","
			}
			undoSql = undoSql[0 : strings.Count(undoSql, "")-2]
			undoSql += " WHERE " + primaryK + "='" + primaryV + "'"
		}
		util.LogNotice(traceId, "undo sql is "+undoSql)

		_, err = oneDB.Exec(undoSql)
		if err != nil {
			return err
		}
	}
	return nil
}

func handleGrError(code int, traceId string, err error, info string) *proto.GlobalRollbackReply {
	if info != "" {
		err = errors.New(err.Error() + " Info:" + info)
	}
	util.LogError(traceId, err)
	return &proto.GlobalRollbackReply{
		ReplyInfo: reply(code, err.Error()),
		TraceId:   traceId,
	}
}
