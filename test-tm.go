package main

import (
	"goseata/proto"
	"goseata/rm"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gofrs/uuid"
)

func main() {
	i := 1
	var wg sync.WaitGroup
	for {
		if i == 11 {
			break
		}
		println("one go.")
		wg.Add(1)
		go run(i, &wg)

		i++
	}
	wg.Wait()
	println("all done.")
}

func run(index int, wg *sync.WaitGroup) {
	defer wg.Done()

	info := &proto.ServiceInfo{
		Appid: "100",
		Name:  "rm1",
	}
	var infoList []*proto.ServiceInfo
	infoList = append(infoList, info)

	rmInstance := rm.New()

	u3, _ := uuid.NewV4()
	traceId := u3.String()

	timeStart := time.Now().UnixNano() / 1e6
RETRY:
	res, err := rmInstance.DoSQL(traceId, "", "", "begin;")
	if err != nil {
		println("begin error:" + err.Error())
	}

	res, err = rmInstance.DoSQL(res.TraceId, res.Tid, res.Ltid, "select * from `user` where id = 1;")
	if err != nil {
		println("select error:" + err.Error())
		res = rollback(rmInstance, res)
		goto RETRY
	}

	res, err = rmInstance.DoSQL(res.TraceId, res.Tid, res.Ltid, "update `user` set money = money - 1 where id = 2;")
	if err != nil {
		println("update error:" + err.Error())
		res = rollback(rmInstance, res)
		goto RETRY
	}

	res, err = rmInstance.DoSQL(res.TraceId, res.Tid, res.Ltid, "commit;")
	if err != nil {
		println(traceId + "commit error:" + err.Error())
		res = rollback(rmInstance, res)
		goto RETRY
	}

	//res, err = rmInstance.DoSQL(res.TraceId, res.Tid, res.Ltid, "gcommit;")
	//if err != nil {
	//	println("gcommit error:" + err.Error())
	//}

	res, err = rmInstance.DoSQL(res.TraceId, res.Tid, res.Ltid, "grollback;")
	if err != nil {
		println("grollback error:" + err.Error())
	}

	timeEnd := time.Now().UnixNano() / 1e6
	println("go " + strconv.Itoa(index) + " time gap:" + strconv.FormatInt(timeEnd-timeStart, 10))
}

func rollback(rmInstance *rm.Rm, res *rm.RmTrans) *rm.RmTrans {
	res2, err := rmInstance.DoSQL(res.TraceId, res.Tid, res.Ltid, "rollback;")
	if err != nil {
		println("rollback error:" + err.Error())
		os.Exit(1)
	}
	return res2
}
