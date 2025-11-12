package handler

import (
	"context"
	"errors"

	"github.com/glekoz/online-shop_proto/user"
	"github.com/glekoz/online-shop_user/shared/models"
	"github.com/glekoz/online-shop_user/shared/myerrors"
	"github.com/glekoz/online-shop_user/shared/validator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserService struct {
	app AppAPI
	user.UnimplementedUserServer
}

type AppAPI interface {
	Register(ctx context.Context, name, email, barePassword string) (access string, refresh string, err error)
	Login(ctx context.Context, email, barePassword string) (access string, refresh string, err error)
	RequestEmailConfirmation(ctx context.Context, userID string) error
	ConfirmEmail(ctx context.Context, userID, mailtoken string) error

	ParseJWTToken(tokenString string) (models.UserToken, error)
}

func (s *UserService) Register(ctx context.Context, req *user.RegisterUserRequest) (*user.LogRegResponse, error) {
	userreq := models.RegisterUserReq{
		Username: req.GetUsername(),
		Password: req.GetPassword(),
		Email:    req.GetEmail(),
	}
	v := validator.New()
	userreq.Validate(v)
	if !v.Valid() {
		return logRegValidationErrorResponse(v)
	}
	access, refresh, err := s.app.Register(ctx, userreq.Username, userreq.Email, userreq.Password)
	if err != nil {
		if errors.Is(err, myerrors.ErrAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "user with the same email already exists")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &user.LogRegResponse{AccessToken: access, RefreshToken: refresh}, nil
}

func (s *UserService) Login(ctx context.Context, req *user.LoginUserRequest) (*user.LogRegResponse, error) {
	userreq := models.LoginUserReq{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}
	v := validator.New()
	userreq.Validate(v)
	if !v.Valid() {
		return logRegValidationErrorResponse(v)
	}
	access, refresh, err := s.app.Login(ctx, userreq.Email, userreq.Password)
	if err != nil {
		if errors.Is(err, myerrors.ErrNotFound) || errors.Is(err, myerrors.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid email or password")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &user.LogRegResponse{AccessToken: access, RefreshToken: refresh}, nil
}

// этот запрос поступает из личного кабинета, поэтому необходимо сверить айди отправителя и айди запрашиваемого аккаунта
func (s *UserService) SendEmailConfirmation(ctx context.Context, req *user.UserID) (*user.Empty, error) {
	id := req.GetId()
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "user id can not be empty")
	}
	// мб стоит в горутине это отправлять
	err := s.app.RequestEmailConfirmation(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, myerrors.ErrInvalidCredentials):
			return nil, status.Error(codes.InvalidArgument, "user's ID is not acceptable")
		case errors.Is(err, myerrors.ErrForbidden):
			return nil, status.Error(codes.PermissionDenied, "only the user can request email confirmation")
		case errors.Is(err, myerrors.ErrNotFound):
			return nil, status.Error(codes.NotFound, "no user found")
		case errors.Is(err, myerrors.ErrAlreadyExists):
			return nil, status.Error(codes.AlreadyExists, "email already confirmed")
		case errors.Is(err, myerrors.ErrMailSent):
			return nil, status.Error(codes.FailedPrecondition, "confirmation letter has already been sent, check your email")
		default:
			// или codes.Internal?
			return nil, status.Error(codes.Unavailable, "mail server is busy")
		}
	}
	return &user.Empty{}, nil
}

func (s *UserService) ConfirmEmail(ctx context.Context, req *user.ConfirmEmailRequest) (*user.Empty, error) {
	userID := req.GetUserID()
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "user id can not be empty")
	}
	mailToken := req.GetMailToken()
	if mailToken == "" {
		return nil, status.Error(codes.InvalidArgument, "mail token can not be empty")
	}
	err := s.app.ConfirmEmail(ctx, userID, mailToken)
	if err != nil {
		switch {
		case errors.Is(err, myerrors.ErrNotFound):
			return nil, status.Error(codes.NotFound, "user not found")
		case errors.Is(err, myerrors.ErrNoMailToken):
			return nil, status.Error(codes.FailedPrecondition, "provided token has been expired or does not exist")
		default:
			return nil, status.Error(codes.Internal, "something goes wrong")
		}
	}
	return &user.Empty{}, nil
}

//GetNewAccessToken
