package error

import (
	"errors"
	"fmt"
	"goseata/proto"
	"goseata/tc/log"
)

type Handler struct {
	traceId string
	code    int
	err     error
	extend  []interface{}
}

func New(traceId string, code int, err error, extends ...interface{}) *Handler {
	if err == nil {
		err = errors.New("none")
	}
	return &Handler{
		traceId: traceId,
		code:    code,
		err:     err,
		extend:  extends,
	}
}

func (h *Handler) ToStr() string {
	str := "TraceID:" + h.traceId + " Error:" + h.err.Error()
	if len(h.extend) != 0 {
		str += " ExtendInfo:" + fmt.Sprint(h.extend)
	}
	return str
}

func (h *Handler) ToReplyInfo() *proto.ReplyInfo {
	return &proto.ReplyInfo{
		Code:    int32(h.code),
		Message: h.ToStr(),
	}
}

func (h *Handler) ToLog() {
	log.HandlerInstance.SendError(h.ToStr())
}
