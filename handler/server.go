package handler

import (
	"fmt"
	"net"

	"google.golang.org/grpc"

	"github.com/glekoz/online-shop_proto/user"
)

func NewServer(app AppAPI) *UserService {
	return &UserService{app: app}
}

func (us *UserService) RunServer(port int) error {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	serv := grpc.NewServer(
		(grpc.ChainUnaryInterceptor(
			us.RequireAuthInterceptor,
			us.RequireNoAuthInterceptor,
		)),
	)
	user.RegisterUserServer(serv, us)
	return serv.Serve(listen)
}
