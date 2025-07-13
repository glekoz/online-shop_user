package grpc

import (
	"log"
	"net"

	"github.com/Gleb988/online-shop_proto/protouser"
	"github.com/Gleb988/online-shop_user/internal/ports"
	"google.golang.org/grpc"
)

type grpcService struct {
	api ports.AppAPI
	protouser.UnimplementedUserServer
}

func NewGrpcService(api ports.AppAPI) *grpcService {
	return &grpcService{
		api: api,
	}
}

func (s *grpcService) Run() {
	listen, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("failed to listen port") // вообще говоря, надо вызвать
		// логгер из структуры бизнес-логики
	}
	grpcServer := grpc.NewServer()
	protouser.RegisterUserServer(grpcServer, s)
	err = grpcServer.Serve(listen)
	if err != nil {
		log.Fatal("failed to serve grpc on port 8080")
	}
}
