package client

import (
	"errors"
	"goseata/proto"
)

func GCommit(path *proto.Path,tid string,traceId string) error {

	path.Infos = append(path.Infos,GetServiceInfo())

	ConnectTc()
	r, err := Tc.client.GlobalCommit(Tc.ctx, &proto.GlobalCommitRequest{
		RequestPath:path,
		Tid:tid,
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
