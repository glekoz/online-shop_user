package handler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/glekoz/online-shop_proto/user"
	"github.com/glekoz/online-shop_user/shared/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const AuthKey = "authorization"
const IPAddress = "ipaddress"

func (us *UserService) RequireNoAuthInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	switch info.FullMethod {
	case user.User_Register_FullMethodName:
		if err := readNoValueFromMD(ctx, AuthKey, "user must not be authenticated", codes.PermissionDenied); err != nil {
			us.logger.InfoContext(ctx, "user must not be authenticated")
			return nil, err
		}
	case user.User_Login_FullMethodName:
		if err := readNoValueFromMD(ctx, AuthKey, "user must not be authenticated", codes.PermissionDenied); err != nil {
			us.logger.InfoContext(ctx, "user must not be authenticated")
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
			us.logger.InfoContext(ctx, "user must be authenticated")
			return nil, status.Error(codes.Unauthenticated, "user must be authenticated")
		}
		if len(md.Get(AuthKey)) > 1 {
			us.logger.InfoContext(ctx, "client provides too much tokens")
			return nil, status.Error(codes.InvalidArgument, "client provides too much tokens")
		}
		token := md.Get(AuthKey)[0]
		u, err := us.app.ParseJWTToken(token)
		if err != nil || u.ID == "" {
			us.logger.InfoContext(ctx, "client provides invalid token")
			return nil, status.Error(codes.Unauthenticated, "client provides invalid token")
		}
		ctx = logger.WithUserID(ctx, u.ID)
	}

	resp, err := handler(ctx, req)

	return resp, err
}

func (us *UserService) PanicRecoverer(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	defer func() {
		if erro := recover(); erro != nil {
			us.logger.ErrorContext(ctx, fmt.Sprintf("panic recovered: %s", erro))
			resp = nil
			err = status.Error(codes.Internal, "Server Internal Error")
		}
	}()
	resp, err = handler(ctx, req)

	return resp, err
}

// rate limiter будет первым, чтобы извлечь из метаданных контекста айпи адрес
func (us *UserService) RateLimiter(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	ip, err := readExactlyOneValueFromMD(ctx, IPAddress, "no IP address provided", codes.Internal)
	if err != nil {
		us.logger.Error("no IP address provided")
		return nil, err
	}
	ctx = logger.WithMethod(ctx, info.FullMethod)
	ctx = logger.WithIPAddress(ctx, ip)
	err = us.rl.Allow(ip)
	if err != nil {
		us.logger.InfoContext(ctx, err.Error())
		return nil, status.Error(codes.ResourceExhausted, err.Error())
	}

	resp, err := handler(ctx, req)

	return resp, err
}

func (us *UserService) TimeCounter(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	start := time.Now()
	us.logger.InfoContext(ctx, "incoming request", slog.String("start time", start.Format("02-01-2006 15:04:05")))

	resp, err := handler(ctx, req)
	if err != nil {
		us.logger.InfoContext(ctx, "request completed successfully", slog.String("time spent", time.Since(start).String()))
	} else {
		us.logger.InfoContext(ctx, "request completed with error", slog.String("time spent", time.Since(start).String()))
	}

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
