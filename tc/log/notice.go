package log

import (
	"fmt"
	"goseata/proto"
)

type Notice struct {
	traceId string
	extend  []interface{}
}

func New(traceId string, extends ...interface{}) *Notice {
	return &Notice{
		traceId: traceId,
		extend:  extends,
	}
}

func (h *Notice) ToStr() string {
	str := "TraceID:" + h.traceId
	if len(h.extend) != 0 {
		str += " ExtendInfo:" + fmt.Sprint(h.extend)
	}
	return str
}

func (h *Notice) ToReplyInfo() *proto.ReplyInfo {
	return &proto.ReplyInfo{
		Code:    0,
		Message: "success",
	}
}

func (h *Notice) ToLog() {
	HandlerInstance.SendNotice(h.ToStr())
}
