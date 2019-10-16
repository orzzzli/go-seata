package util

import "goseata/proto"

func Trace(traceId string, funcName string, path *proto.Path)  {
	infos := path.Infos
	fullStr := "func: "+funcName+" "
	for _, info := range infos {
		fullStr += info.Appid+"."+info.Name+"->"
	}
	fullStr = string(fullStr[:len(fullStr) - 2])
	LogTrace(traceId,fullStr)
}
