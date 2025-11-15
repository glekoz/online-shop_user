package handler

import (
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"

	"github.com/glekoz/online-shop_proto/user"
)

type UserService struct {
	user.UnimplementedUserServer
	app AppAPI

	logger *slog.Logger
	rl     *RateLimiter
}

func NewServer(app AppAPI, l *slog.Logger) *UserService {
	return &UserService{app: app, logger: l, rl: NewRateLimiter()}
}

func (us *UserService) RunServer(port int) error {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	serv := grpc.NewServer(
		(grpc.ChainUnaryInterceptor(
			us.RateLimiter,
			us.TimeCounter,
			us.RequireAuthInterceptor,
			us.RequireNoAuthInterceptor,
			us.PanicRecoverer,
		)),
	)
	user.RegisterUserServer(serv, us)
	return serv.Serve(listen)
}
