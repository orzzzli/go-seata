package action

import (
	"context"
	"encoding/json"
	"errors"
	"goseata/proto"
	"goseata/tc/consts"
	error2 "goseata/tc/error"
	"goseata/tc/lock"
	"goseata/tc/log"
	"goseata/tc/model"
	"goseata/tc/mysql"
	"strings"

	"github.com/orzzzli/orzconfiger"
)

func (s *TcServer) GlobalRollback(ctx context.Context, req *proto.GlobalRollbackRequest) (*proto.GlobalRollbackReply, error) {
	//打印trace
	log.New(req.TraceId, "globalRollback", req.RequestPath).ToLog()

	transactionDB, find := mysql.DBPoolsInstance.GetPool("db-transaction")
	if !find {
		return handleGrError(req.TraceId, consts.DBError, nil, "transaction db not found.")
	}

	var localTransactions []*model.LocalTransaction
	err := transactionDB.Select(&localTransactions, "SELECT * FROM `local_transaction` where tid = ?", req.Tid)
	if err != nil {
		return handleGrError(req.TraceId, consts.DBError, err, "select local transaction error.", req.Tid)
	}

	for _, v := range localTransactions {
		appid := v.Appid
		connect, find := orzconfiger.GetString(appid, "connect")
		if !find {
			return handleGrError(req.TraceId, consts.ConfigError, nil, "get config connect error.", appid)
		}

		log.New(req.TraceId, "globalRollback.begin rollback", connect, req.Tid).ToLog()

		err := rollback(req.TraceId, connect, req.Tid)
		if err != nil {
			return handleGrError(req.TraceId, consts.BusinessError, err, "rollback error.", connect, req.Tid)
		}
		//更新分支事务状态
		_, err = transactionDB.Exec("update `local_transaction` set status = ? where id = ?", model.LocalTransactionStatusRollbacked, v.Id)
		if err != nil {
			return handleGrError(req.TraceId, consts.DBError, err, "update local transaction status error.", v.Id)
		}

		log.New(req.TraceId, "globalRollback.rollback success", connect, req.Tid).ToLog()
	}

	//清除锁
	appid, find := orzconfiger.GetString("service", "appid")
	if !find {
		return handleGrError(req.TraceId, consts.ConfigError, nil, "get config appid error.", appid)
	}
	lockManager := lock.New(appid, req.Tid, nil)
	err = lockManager.ClearLocks()
	if err != nil {
		return handleGrError(req.TraceId, consts.LockError, err, "clear lock error.", lockManager.ToStr())
	}
	log.New(req.TraceId, "globalRollback.clear lock success.", lockManager.ToStr()).ToLog()

	return &proto.GlobalRollbackReply{
		ReplyInfo: &proto.ReplyInfo{
			Code:    0,
			Message: "success",
		},
		TraceId: req.TraceId,
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

		log.New(traceId, "globalRollback.rollback", undoSql).ToLog()

		_, err = oneDB.Exec(undoSql)
		if err != nil {
			return err
		}
	}
	return nil
}

func handleGrError(traceId string, code int, err error, extends ...interface{}) (*proto.GlobalRollbackReply, error) {
	errorObj := error2.New(traceId, code, err, extends)
	errorObj.ToLog()
	return &proto.GlobalRollbackReply{
		ReplyInfo: errorObj.ToReplyInfo(),
		TraceId:   traceId,
	}, nil
}
