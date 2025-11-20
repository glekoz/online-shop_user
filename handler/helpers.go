package handler

import (
	"context"
	"errors"

	"github.com/glekoz/online-shop_proto/user"
	"github.com/glekoz/online-shop_user/app"
	"github.com/glekoz/online-shop_user/shared/logger"
	"github.com/glekoz/online-shop_user/shared/validator"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func badRequestResponse(reason string, metadata map[string]string) error {
	if len(metadata) < 1 {
		return status.Error(codes.InvalidArgument, "Bad Request")
	}
	st := status.New(codes.InvalidArgument, reason)
	fields := make([]*errdetails.BadRequest_FieldViolation, 0, len(metadata))
	for k, v := range metadata {
		fields = append(fields, &errdetails.BadRequest_FieldViolation{
			Field:       k,
			Description: v,
		})
	}
	details := errdetails.BadRequest{
		FieldViolations: fields,
	}
	st, err := st.WithDetails(&details)
	if err != nil {
		return status.Error(codes.Internal, "детали не добавились")
	}
	return st.Err()
}

func logRegBadRequestResponse(v *validator.Validator) (*user.LogRegResponse, error) {
	err := badRequestResponse("validation of provided credentials failed", v.Errors)
	return nil, err
}

func (us *UserService) handleError(ctx context.Context, err error, args ...any) error {
	switch {
	case errors.Is(err, app.ErrUserAlreadyExists):
		us.logger.InfoContext(ctx, app.ErrUserAlreadyExists.Error(), args...)
		return status.Error(codes.AlreadyExists, "user with the same email already exists")
	case errors.Is(err, app.ErrInvalidCredentials):
		us.logger.InfoContext(ctx, app.ErrInvalidCredentials.Error(), args...)
		return status.Error(codes.Unauthenticated, "wrong email or password")
	case errors.Is(err, app.ErrNoRUID):
		us.logger.InfoContext(logger.ErrorCtx(ctx, err), app.ErrNoRUID.Error(), args...)
		return status.Error(codes.Unauthenticated, "user is not authenticated")
	case errors.Is(err, app.ErrRUIDneID):
		us.logger.InfoContext(logger.ErrorCtx(ctx, err), app.ErrRUIDneID.Error(), args...)
		return status.Error(codes.PermissionDenied, "only the user can request email confirmation")
	case errors.Is(err, app.ErrUserNotFound):
		us.logger.InfoContext(logger.ErrorCtx(ctx, err), app.ErrUserNotFound.Error(), args...)
		return status.Error(codes.NotFound, "no user found")
	case errors.Is(err, app.ErrEmailAlreadyConfirmed):
		us.logger.InfoContext(logger.ErrorCtx(ctx, err), app.ErrEmailAlreadyConfirmed.Error(), args...)
		return status.Error(codes.AlreadyExists, "email already confirmed")
	case errors.Is(err, app.ErrMsgAlreadySent):
		us.logger.InfoContext(logger.ErrorCtx(ctx, err), app.ErrMsgAlreadySent.Error(), args...)
		return status.Error(codes.FailedPrecondition, "confirmation letter has already been sent, check your email")
	case errors.Is(err, app.ErrWrongMailToken):
		us.logger.InfoContext(logger.ErrorCtx(ctx, err), app.ErrWrongMailToken.Error(), args...)
		return status.Error(codes.FailedPrecondition, "provided token has been expired or does not exist")
	default:
		us.logger.ErrorContext(logger.ErrorCtx(ctx, err), "unexpected error", "error", err.Error())
		return status.Error(codes.Internal, "something went wrong")
	}
}
