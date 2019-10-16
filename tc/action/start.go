package action
//
//import (
//	"context"
//	"errors"
//	"goseata/proto"
//	"goseata/tc/mysql"
//	"goseata/util"
//	"strconv"
//)
//
//
//func (s *TcServer) Start(ctx context.Context, req *proto.StartRequest) (*proto.StartReply, error) {
//	//打印trace
//	util.Trace("start",req.RequestPath)
//
//	firstService := req.RequestPath.Infos[0]
//
//	transactionDB, find := mysql.DBPoolsInstance.GetPool("db-transaction")
//	if !find {
//		err := errors.New("transaction db not found")
//		util.LogError(err)
//		return &proto.StartReply{
//			ReplyInfo:reply(1,err.Error()),
//		}, nil
//	}
//
//	//插入db
//	result,err := transactionDB.Exec("INSERT INTO `transaction`(appid,name)VALUES (?,?)",firstService.Appid,firstService.Name)
//	if err != nil {
//		util.LogError(err)
//		return &proto.StartReply{
//			ReplyInfo:reply(1,"insert error"),
//		}, nil
//	}
//	tid,_ := result.LastInsertId()
//	tidStr := strconv.FormatInt(tid,10)
//
//	return &proto.StartReply{
//		ReplyInfo:reply(0,"success"),
//		Tid:tidStr,
//	}, nil
//}
//
