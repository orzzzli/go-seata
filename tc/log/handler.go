package log

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/orzzzli/orzconfiger"
)

type Handler struct {
	noticeChan chan string
	errorChan  chan string
	noticeFile *os.File
	errorFile  *os.File
}

var HandlerInstance *Handler

func InitLogHandler() {
	buffSize, ok := orzconfiger.GetInt("log", "bufferLen")
	if !ok {
		log.Fatalln("get log config bufferLen error.")
	}

	logPath, find := orzconfiger.GetString("log", "path")
	if !find {
		log.Fatalln("get log config path error.")
	}
	//init notice file handler
	noticeFullPath := logPath + "notice.log"
	noticeF, err := os.OpenFile(noticeFullPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("open notice log file error. full path is " + noticeFullPath)
	}
	//init error file handler
	errorFullPath := logPath + "error.log"
	errorF, err := os.OpenFile(errorFullPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("open error log file error. full path is " + errorFullPath)
	}

	//init channels
	nc := make(chan string, buffSize)
	ec := make(chan string, buffSize)
	HandlerInstance = &Handler{
		noticeChan: nc,
		noticeFile: noticeF,
		errorChan:  ec,
		errorFile:  errorF,
	}

	noticeGoNumbers, ok := orzconfiger.GetInt("log", "noticeWriter")
	if !ok {
		log.Fatalln("get log config noticeWriter error.")
	}
	errorGoNumbers, ok := orzconfiger.GetInt("log", "errorWriter")
	if !ok {
		log.Fatalln("get log config errorWriter error.")
	}

	i := 0
	for ; i < noticeGoNumbers; i++ {
		go HandlerInstance.ReadNotice()
	}
	println(strconv.Itoa(i+1) + " notice writer start.")

	i = 0
	for ; i < errorGoNumbers; i++ {
		go HandlerInstance.ReadError()
	}
	println(strconv.Itoa(i+1) + " error writer start.")
}

func (h *Handler) SendNotice(log string) {
	h.noticeChan <- log
}

func (h *Handler) SendError(log string) {
	h.errorChan <- log
}

func (h *Handler) ReadNotice() {
	for {
		logStr := <-h.noticeChan
		curTimeStr := time.Now().Format("2006-01-02.15:04:05")
		logStr = "Notice.Time:" + curTimeStr + " " + logStr + "\n"
		h.WriteNotice(logStr)
	}
}

func (h *Handler) ReadError() {
	for {
		logStr := <-h.errorChan
		curTimeStr := time.Now().Format("2006-01-02.15:04:05")
		logStr = "Error.Time:" + curTimeStr + " " + logStr + "\n"
		h.WriteError(logStr)
	}
}

func (h *Handler) WriteNotice(log string) {
	h.noticeFile.Write([]byte(log))
}

func (h *Handler) WriteError(log string) {
	h.errorFile.Write([]byte(log))
}
