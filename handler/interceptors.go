package handler

import (
	"context"
	"fmt"

	"github.com/glekoz/online-shop_proto/user"
	"github.com/glekoz/online-shop_user/shared/vars"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const AuthKey = "authorization"

// тут можно опционально вставлять RUID в контекст
func (us *UserService) RequireNoAuthInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	switch info.FullMethod {
	case user.User_Register_FullMethodName:
		if err := readMDNoAuth(ctx); err != nil {
			return nil, err
		}
	case user.User_Login_FullMethodName:
		if err := readMDNoAuth(ctx); err != nil {
			return nil, err
		}
	default:
	}

	resp, err := handler(ctx, req)

	return resp, err
}

func (us *UserService) RequireAuthInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	switch info.FullMethod {
	case user.User_Register_FullMethodName:
	case user.User_Login_FullMethodName:
	case user.User_GetNewAccessToken_FullMethodName: // для этого метода неважно наличие или отсутствие токена
	case user.User_GetRSAPublicKey_FullMethodName: // для этого тоже
	default:
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok || len(md.Get(AuthKey)) < 1 {
			return nil, status.Error(codes.Unauthenticated, "user must be authenticated")
		}
		if len(md.Get(AuthKey)) > 1 {
			return nil, status.Error(codes.InvalidArgument, "client provides too much tokens")
		}
		token := md.Get(AuthKey)[0]
		u, err := us.app.ParseJWTToken(token)
		if err != nil || u.ID == "" {
			// мб, залогировать
			fmt.Println(err)
			return nil, status.Error(codes.Unauthenticated, "client provides invalid token")
		}
		ctx = context.WithValue(ctx, vars.ContextKeyRequestUserID, u.ID)
	}

	resp, err := handler(ctx, req)

	return resp, err
}

func readMDNoAuth(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if len(md.Get(AuthKey)) > 0 {
			return status.Error(codes.FailedPrecondition, "user must not be authenticated")
		}
	}
	return nil
}
