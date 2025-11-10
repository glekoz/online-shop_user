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
	Register(ctx context.Context, email, password string) (string, error)
	Login(ctx context.Context, email, password string) (string, error)
}

func (s *UserService) Register(ctx context.Context, req *user.RegisterUserRequest) (*user.RegisterUserResponse, error) {
	userreq := models.RegisterUserReq{
		Username: req.GetUsername(),
		Password: req.GetPassword(),
		Email:    req.GetEmail(),
	}
	v := validator.New()
	userreq.Validate(v)
	if !v.Valid() {
		return regValidationErrorResponse(v)
	}
	token, err := s.app.Register(ctx, userreq.Username, userreq.Password)
	if err != nil {
		if errors.Is(err, myerrors.ErrAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "user with the same username or email already exists")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &user.RegisterUserResponse{Token: token}, nil
}

func (s *UserService) Login(ctx context.Context, req *user.LoginUserRequest) (*user.LoginUserResponse, error) {
	userreq := models.LoginUserReq{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}
	v := validator.New()
	userreq.Validate(v)
	if !v.Valid() {
		err := validationErrorResponse("Validation", "user", v.Errors)
		return nil, err
	}
	token, err := s.app.Login(ctx, userreq.Email, userreq.Password)
	if err != nil {
		if errors.Is(err, myerrors.ErrNotFound) || errors.Is(err, myerrors.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid email or password")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &user.LoginUserResponse{Token: token}, nil
}
