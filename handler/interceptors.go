package handler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/glekoz/online-shop_proto/user"
	"github.com/glekoz/online-shop_user/shared/vars"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const AuthKey = "authorization"
const IPAddress = "ipaddress"

// us.logger.Info("incoming request", slog.String("resource", info.FullMethod))
// тут можно опционально вставлять RUID в контекст
func (us *UserService) RequireNoAuthInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	switch info.FullMethod {
	case user.User_Register_FullMethodName:
		if err := readNoValueFromMD(ctx, AuthKey, "user must not be authenticated", codes.PermissionDenied); err != nil {
			return nil, err
		}
	case user.User_Login_FullMethodName:
		if err := readNoValueFromMD(ctx, AuthKey, "user must not be authenticated", codes.PermissionDenied); err != nil {
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

func (us *UserService) PanicRecoverer(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	defer func() {
		if erro := recover(); erro != nil {
			resp = nil
			err = status.Error(codes.Internal, fmt.Sprintf("panic recovered: %s", erro))
		}
	}()
	resp, err = handler(ctx, req)

	return resp, err
}

// rate limiter будет первым, чтобы извлечь из метаданных контекста айпи адрес
func (us *UserService) RateLimiter(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	ip, err := readExactlyOneValueFromMD(ctx, IPAddress, "no IP address provided", codes.Internal)
	if err != nil {
		us.logger.Error(err.Error())
		return nil, err
	}
	err = us.rl.Allow(ip)
	if err != nil {
		us.logger.Info(err.Error(), slog.String("ip address", ip))
		return nil, status.Error(codes.ResourceExhausted, err.Error())
	}
	ctx = context.WithValue(ctx, vars.ContextKeyRequestIPAddress, ip)

	resp, err := handler(ctx, req)

	return resp, err
}

func (us *UserService) TimeCounter(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	start := time.Now()
	ip, _ := ctx.Value(vars.ContextKeyRequestIPAddress).(string)
	us.logger.Info("incoming request", slog.String("ip address", ip), slog.String("start time", start.Format("02-01-2006 15:04:05")))

	resp, err := handler(ctx, req)

	us.logger.Info("request completed", slog.String("ip address", ip), slog.String("time spent", time.Since(start).String()))

	return resp, err
}

// но если передать codes.OK, то ошибка будет nil
func readExactlyOneValueFromMD(ctx context.Context, key, msg string, code codes.Code) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		// internal, потому что это моя вина, если я не положил в метаданные информацию на фронте
		return "", status.Error(codes.Internal, "no metadata")
	}
	data := md.Get(key)
	if len(data) != 1 || data[0] == "" {
		return "", status.Error(code, msg)
	}
	return data[0], nil
}

func readNoValueFromMD(ctx context.Context, key, msg string, code codes.Code) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if len(md.Get(key)) > 0 {
			return status.Error(code, msg)
		}
	}
	return nil
}
