package util

import "log"

func LogNotice(traceId string, message string) {
	log.Println("Trace: " + traceId + ". Notice: " + message)
}

func LogError(traceId string, err error) {
	log.Println("Trace: " + traceId + " Error: " + err.Error())
}

func LogTrace(traceId string, trace string) {
	log.Println("Trace: " + traceId + " TracePath: " + trace)
}
