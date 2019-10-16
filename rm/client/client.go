package client

import (
	"context"
	"google.golang.org/grpc"
	"goseata/proto"
	"log"
	"time"
)

const (
	TCDriver = "localhost:50051"
)

type TCClient struct {
	client proto.TcServerClient
	ctx context.Context
}

var Tc TCClient

func ConnectTc() {
	conn, err := grpc.Dial(TCDriver, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	Tc.client = proto.NewTcServerClient(conn)
	Tc.ctx, _ = context.WithTimeout(context.Background(), time.Second*300)
}

func GetServiceInfo() *proto.ServiceInfo {
	return &proto.ServiceInfo{
		Appid:"100",
		Name:"tm1",
	}
}