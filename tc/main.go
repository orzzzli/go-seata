package main

import (
	"goseata/proto"
	"goseata/tc/action"
	log2 "goseata/tc/log"
	"goseata/tc/mysql"
	"goseata/tc/redis"
	"log"
	"net"

	"github.com/orzzzli/orzconfiger"

	"google.golang.org/grpc"
)

func main() {
	//init configer
	orzconfiger.InitConfiger("config.ini")

	//init dbs
	err := mysql.DBPoolsInstance.InitPool("db-transaction")
	if err != nil {
		log.Fatalf("init db error: %v", err)
	}
	err = mysql.DBPoolsInstance.InitPool("db-user")
	if err != nil {
		log.Fatalf("init db error: %v", err)
	}

	//init redis
	err = redis.InitRedis("redis-common")
	if err != nil {
		log.Fatalf("init redis error: %v", err)
	}

	//init logger
	log2.InitLogHandler()

	//init server
	host, find := orzconfiger.GetString("tcp", "host")
	if !find {
		log.Fatalf("failed to load config host")
	}
	port, find := orzconfiger.GetString("tcp", "port")
	if !find {
		log.Fatalf("failed to load config port")
	}

	address := host + ":" + port
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Println("server run. listen " + address)

	s := grpc.NewServer()
	proto.RegisterTcServerServer(s, &action.TcServer{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
