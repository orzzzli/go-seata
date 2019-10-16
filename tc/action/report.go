package action

import (
	"context"
	"errors"
	"goseata/proto"
	"goseata/tc/mysql"
	"goseata/util"
)

func (s *TcServer) Report(ctx context.Context, req *proto.ReportRequest) (*proto.ReportReply, error) {
	//打印trace
	util.Trace(req.TraceId, "report", req.RequestPath)

	status := req.Status

	transactionDB, find := mysql.DBPoolsInstance.GetPool("db-transaction")
	if !find {
		err := errors.New("transaction db not found")
		return handleRepError(1000, req.TraceId, err, ""), nil
	}

	//更新分支事务状态
	_, err := transactionDB.Exec("update `local_transaction` set status = ? where id = ?", status, req.Ltid)
	if err != nil {
		return handleRepError(1100, req.TraceId, err, "update local transaction status error"), nil
	}

	//触发全局回滚
	if status == proto.LocalTransactionStatus_ROLLBACKED {
		util.LogNotice(req.TraceId, "trigger global rollback. tid is "+req.Tid)
		greply, err := s.GlobalRollback(ctx, &proto.GlobalRollbackRequest{
			RequestPath: req.RequestPath,
			Tid:         req.Tid,
			TraceId:     req.TraceId,
		})
		if err != nil {
			return handleRepError(1200, req.TraceId, err, "global rollback error"), nil
		}
		if greply.ReplyInfo.Code != 0 {
			err = errors.New(greply.ReplyInfo.Message)
			return handleRepError(int(greply.ReplyInfo.Code), greply.TraceId, err, ""), nil
		}
	}

	return &proto.ReportReply{
		ReplyInfo: reply(0, "success"),
		TraceId:   req.TraceId,
	}, nil
}

func handleRepError(code int, traceId string, err error, info string) *proto.ReportReply {
	if info != "" {
		err = errors.New(err.Error() + " Info:" + info)
	}
	util.LogError(traceId, err)
	return &proto.ReportReply{
		ReplyInfo: reply(code, err.Error()),
		TraceId:   traceId,
	}
}
