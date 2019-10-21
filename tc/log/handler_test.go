package log

import (
	"testing"
	"time"

	"github.com/orzzzli/orzconfiger"
)

func TestHandler_SendNotice(t *testing.T) {
	orzconfiger.InitConfiger("/Users/Orz/Desktop/goProjects/go-seata/tc/config.ini")
	InitLogHandler()
	HandlerInstance.SendNotice("testa")
	HandlerInstance.SendNotice("testb")
	HandlerInstance.SendNotice("testc")
	HandlerInstance.SendError("testcccc")
	HandlerInstance.SendError("testdddd")
	HandlerInstance.SendError("testeeee")
	HandlerInstance.SendError("testffff")
	time.Sleep(time.Second * 5)
}

/*
	100000times	     12432 ns/op
*/
func BenchmarkHandler_SendNotice(b *testing.B) {
	orzconfiger.InitConfiger("/Users/Orz/Desktop/goProjects/go-seata/tc/config.ini")
	InitLogHandler()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HandlerInstance.SendNotice("testa")
		HandlerInstance.SendError("testcccc")
	}
}
