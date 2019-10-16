package client

import (
	"errors"
	"goseata/proto"
	"goseata/rm/lock"
	"strings"
)

func Register(path *proto.Path, tid string, traceId string) (string, bool, error) {

	serviceInfo := GetServiceInfo()
	path.Infos = append(path.Infos, serviceInfo)

	locks, err := lock.GetLocalLocks(tid)
	if err != nil {
		return "", false, nil
	}
	var lockList []*proto.Lock
	for _, v := range locks {
		tempSlice := strings.Split(v, "|")
		oneLock := &proto.Lock{
			Connection: tempSlice[0],
			Database:   tempSlice[1],
			Table:      tempSlice[2],
			PrimaryK:   tempSlice[3],
			PrimaryV:   tempSlice[4],
		}
		lockList = append(lockList, oneLock)
	}

	request := &proto.RegisterRequest{
		RequestPath: path,
		Tid:         tid,
		Locks:       lockList,
		TraceId:     traceId,
	}

	ConnectTc()
	r, err := Tc.client.Register(Tc.ctx, request)
	if err != nil {
		return "", false, err
	}
	if r.ReplyInfo.Code != 0 {
		return "", false, errors.New(r.ReplyInfo.Message)
	}
	if !r.Lock {
		return "", false, nil
	}
	return r.Ltid, true, nil
}
