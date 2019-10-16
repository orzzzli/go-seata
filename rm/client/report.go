package client

import (
	"errors"
	"goseata/proto"
)

func Report(path *proto.Path, tid string, ltid string, status proto.LocalTransactionStatus,traceId string) error {

	path.Infos = append(path.Infos,GetServiceInfo())

	ConnectTc()
	r, err := Tc.client.Report(Tc.ctx, &proto.ReportRequest{
		RequestPath:path,
		Tid:tid,
		Ltid:ltid,
		Status:status,
		TraceId:traceId,
	})
	if err != nil {
		return err
	}
	if r.ReplyInfo.Code != 0 {
		return errors.New(r.ReplyInfo.Message)
	}
	return nil
}

