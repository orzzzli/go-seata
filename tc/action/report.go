package action

import (
	"context"
	"goseata/proto"
	"goseata/tc/consts"
	error2 "goseata/tc/error"
	"goseata/tc/log"
	"goseata/tc/mysql"
)

func (s *TcServer) Report(ctx context.Context, req *proto.ReportRequest) (*proto.ReportReply, error) {
	//打印trace
	log.New(req.TraceId, "report", req.RequestPath).ToLog()

	status := req.Status

	transactionDB, find := mysql.DBPoolsInstance.GetPool("db-transaction")
	if !find {
		return handleRepError(req.TraceId, consts.DBError, nil, "transaction db not found.")
	}

	//更新分支事务状态
	_, err := transactionDB.Exec("update `local_transaction` set status = ? where id = ?", status, req.Ltid)
	if err != nil {
		return handleRepError(req.TraceId, consts.DBError, err, "update local transaction status error.", status, req.Ltid)
	}

	//触发全局回滚
	if status == proto.LocalTransactionStatus_ROLLBACKED {
		log.New(req.TraceId, "report", "trigger global rollback.", req.Tid).ToLog()

		greply, err := s.GlobalRollback(ctx, &proto.GlobalRollbackRequest{
			RequestPath: req.RequestPath,
			Tid:         req.Tid,
			TraceId:     req.TraceId,
		})
		if err != nil {
			return handleRepError(req.TraceId, consts.GRPCError, err, "global rollback error.", req.RequestPath, req.Tid, req.TraceId)
		}
		if greply.ReplyInfo.Code != 0 {
			return handleRepError(req.TraceId, consts.BusinessError, nil, greply.ReplyInfo.Message, greply)
		}
	}
	log.New(req.TraceId, "report", "success.", req.Tid, status).ToLog()

	return &proto.ReportReply{
		ReplyInfo: &proto.ReplyInfo{
			Code:    0,
			Message: "success",
		},
		TraceId: req.TraceId,
	}, nil
}

func handleRepError(traceId string, code int, err error, extends ...interface{}) (*proto.ReportReply, error) {
	errorObj := error2.New(traceId, code, err, extends)
	errorObj.ToLog()
	return &proto.ReportReply{
		ReplyInfo: errorObj.ToReplyInfo(),
		TraceId:   traceId,
	}, nil
}
