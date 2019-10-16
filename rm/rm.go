package rm

import (
	"database/sql"
	"encoding/json"
	"errors"
	"goseata/proto"
	"goseata/rm/client"
	"goseata/rm/lock"
	"goseata/rm/mysql"
	"goseata/rm/proxy"
	"goseata/util"

	uuid2 "github.com/gofrs/uuid"

	"github.com/jmoiron/sqlx"
)

var RmInstance *Rm

type Rm struct {
	tx     *sql.Tx
	dbConn *sqlx.DB
}

type RmTrans struct {
	TraceId string
	Tid     string
	Ltid    string
	DbRes   interface{}
}

func New() *Rm {
	RmInstance = &Rm{
		dbConn: mysql.DBPool,
		tx:     nil,
	}
	return RmInstance
}

func (r *Rm) DoSQL(traceId string, tid string, ltid string, sql string) (trans *RmTrans, err error) {
	util.LogNotice(traceId, "request sql is "+sql)

	message := &RmTrans{
		Tid:     tid,
		Ltid:    ltid,
		TraceId: traceId,
		DbRes:   nil,
	}

	sqlProxy, err := proxy.New(tid, sql)
	if err != nil {
		return message, err
	}

	//开启事务，并且全局事务id为""
	if sqlProxy.SQLType == proxy.SQLBegin && sqlProxy.Tid == "" {
		util.LogNotice(traceId, "start global transaction.")
		//生成全局事务id
		uuid, err := uuid2.NewV4()
		if err != nil {
			return message, err
		}
		sqlProxy.Tid = uuid.String()
		message.Tid = sqlProxy.Tid
		util.LogNotice(traceId, "new tid is "+sqlProxy.Tid)
	}

	//commit前请求锁
	if sqlProxy.SQLType == proxy.SQLCommit {
		getLock := false
		//请求锁
		ltid, getLock, err = client.Register(&proto.Path{}, tid, traceId)
		if err != nil {
			return message, err
		}
		//未请求到锁
		if !getLock {
			err = errors.New("get lock fail")
			return message, err
		}
	}

	res, err := r.Execute(sqlProxy)
	if err != nil {
		return message, err
	}
	message.DbRes = res

	//commit/rollback后report
	if sqlProxy.SQLType == proxy.SQLCommit || sqlProxy.SQLType == proxy.SQLRollback {
		branchStatus := proto.LocalTransactionStatus_COMMITED
		if sqlProxy.SQLType == proxy.SQLRollback {
			branchStatus = proto.LocalTransactionStatus_ROLLBACKED
		}
		//report分支状态
		err = client.Report(&proto.Path{}, tid, ltid, branchStatus, traceId)
		if err != nil {
			return message, err
		}
		//清除本地锁
		err = lock.RmLocalLock(tid)
		if err != nil {
			return message, err
		}
	}

	if sqlProxy.SQLType == proxy.SQLGlobalCommit {
		//全局提交
		err = client.GCommit(&proto.Path{}, tid, traceId)
		if err != nil {
			return message, err
		}
	}

	if sqlProxy.SQLType == proxy.SQLGlobalRollback {
		//全局回滚
		err = client.GRollback(&proto.Path{}, tid, traceId)
		if err != nil {
			return message, err
		}
	}

	return message, nil
}

func (r *Rm) insertTransactionLog(sqlP *proxy.SQLProxy, ids []string, beforeCols []string, afterCols []string) error {
	if len(beforeCols) != len(afterCols) {
		return errors.New("before select and after select not match")
	}
	for index, v := range ids {
		var err error
		if r.tx != nil {
			//插入transaction_log
			_, err = r.tx.Exec("INSERT INTO `transcation_log`(tid,type,before_col,after_col,table_name,primary_key,primary_value,connection_str)VALUES (?,?,?,?,?,?,?,?)", sqlP.Tid, sqlP.SQLType, beforeCols[index], afterCols[index], sqlP.TableName, "id", v, "db-user")
		} else {
			_, err = r.dbConn.Exec("INSERT INTO `transcation_log`(tid,type,before_col,after_col,table_name,primary_key,primary_value,connection_str)VALUES (?,?,?,?,?,?,?,?)", sqlP.Tid, sqlP.SQLType, beforeCols[index], afterCols[index], sqlP.TableName, "id", v, "db-user")
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Rm) Execute(sqlP *proxy.SQLProxy) (interface{}, error) {
	if sqlP.SQLType == proxy.SQLSelect {
		return r.selectFromDB(sqlP.OriginSQL)
	}
	if sqlP.SQLType == proxy.SQLUpdate {
		ids, beforeCols, err := r.buildBeforeImage(sqlP.BuildBeforeSelectSQL())
		if err != nil {
			return nil, err
		}

		err = r.ExecFromFB(sqlP.OriginSQL)
		if err != nil {
			return nil, err
		}

		afterCols, err := r.buildAfterImage(sqlP.BuildAfterSelectSQL())
		if err != nil {
			return nil, err
		}

		err = r.insertTransactionLog(sqlP, ids, beforeCols, afterCols)
		if err != nil {
			return nil, err
		}

		for _, v := range ids {
			err = lock.SetLocalLock(sqlP.Tid, sqlP.Connection, sqlP.Database, sqlP.TableName, "id", v)
			if err != nil {
				return nil, err
			}
		}
		return nil, nil
	}
	if sqlP.SQLType == proxy.SQLBegin {
		var err error
		r.tx, err = r.dbConn.Begin()
		if err != nil {
			return nil, err
		}
	}
	if sqlP.SQLType == proxy.SQLCommit {
		err := r.tx.Commit()
		if err != nil {
			r.tx.Rollback()
			return nil, err
		}
		r.tx = nil
	}
	if sqlP.SQLType == proxy.SQLRollback {
		err := r.tx.Rollback()
		if err != nil {
			r.tx.Rollback()
			return nil, err
		}
		r.tx = nil
	}
	return nil, nil
}

func (r *Rm) selectFromDB(sqlStr string) ([]map[string]string, error) {
	var rows *sql.Rows
	var err error
	if r.tx != nil {
		rows, err = r.tx.Query(sqlStr)
	} else {
		rows, err = r.dbConn.Query(sqlStr)
	}
	if err != nil {
		return nil, err
	}
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	count := len(columns)
	tableData := make([]map[string]string, 0)
	values := make([]interface{}, count)
	temp := make([]interface{}, count)
	for rows.Next() {
		for i := 0; i < count; i++ {
			temp[i] = &values[i]
		}
		err := rows.Scan(temp...)
		if err != nil {
			return nil, err
		}
		oneRow := make(map[string]string)
		for i, key := range columns {
			val := values[i]
			temp, ok := val.([]byte)
			if !ok {
				return nil, errors.New("select value convert error. sql is " + sqlStr)
			}
			oneRow[key] = string(temp)
		}
		tableData = append(tableData, oneRow)
	}
	return tableData, nil
}

func (r *Rm) ExecFromFB(sql string) error {
	var err error
	if r.tx != nil {
		_, err = r.tx.Exec(sql)
	} else {
		_, err = r.dbConn.Exec(sql)
	}
	return err
}

func (r *Rm) buildBeforeImage(beforeSql string) ([]string, []string, error) {
	var ids []string
	var cols []string
	values, err := r.selectFromDB(beforeSql)
	if err != nil {
		return nil, nil, err
	}
	for _, v := range values {
		id, ok := v["id"]
		if !ok {
			return nil, nil, errors.New("select before not found id. sql is " + beforeSql)
		}
		ids = append(ids, id)
		delete(v, "id")
		beforeCol, err := json.Marshal(v)
		if err != nil {
			return nil, nil, err
		}
		cols = append(cols, string(beforeCol))
	}
	return ids, cols, nil
}

func (r *Rm) buildAfterImage(afterSql string) ([]string, error) {
	var cols []string
	values, err := r.selectFromDB(afterSql)
	if err != nil {
		return nil, err
	}
	for _, v := range values {
		delete(v, "id")
		beforeCol, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		cols = append(cols, string(beforeCol))
	}
	return cols, nil
}
