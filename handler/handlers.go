package handler

import (
	"context"
	"errors"

	"github.com/glekoz/online-shop_proto/user"
	"github.com/glekoz/online-shop_user/app"
	"github.com/glekoz/online-shop_user/shared/logger"
	"github.com/glekoz/online-shop_user/shared/models"
	"github.com/glekoz/online-shop_user/shared/myerrors"
	"github.com/glekoz/online-shop_user/shared/validator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AppAPI interface {
	Register(ctx context.Context, name, email, barePassword string) (access string, refresh string, err error)
	Login(ctx context.Context, email, barePassword string) (access string, refresh string, err error)
	RequestEmailConfirmation(ctx context.Context, userID string) error
	ConfirmEmail(ctx context.Context, userID, mailtoken string) error

	ParseJWTToken(tokenString string) (models.UserToken, error)
	CreateJWTToken(userID, name string, isModer, isAdmin, isCore bool) (string, error)
	GetRSAPublicKey() ([]byte, error)
}

func (us *UserService) Register(ctx context.Context, req *user.RegisterUserRequest) (*user.LogRegResponse, error) {
	userreq := models.RegisterUserReq{
		Username: req.GetUsername(),
		Password: req.GetPassword(),
		Email:    req.GetEmail(),
	}
	v := validator.New()
	userreq.Validate(v)
	if !v.Valid() {
		// можно было бы добавить тэги для RegisterUserReq или добавить метод MarshalJSON,
		// но если логировать понадобится в текстовом формате, то чувствительные данные
		// попадут в логи
		us.logger.InfoContext(ctx, "validation failed", "input data", map[string]string{"username": userreq.Username, "email": userreq.Email})
		return logRegBadRequestResponse(v)
	}
	access, refresh, err := us.app.Register(ctx, userreq.Username, userreq.Email, userreq.Password)
	if err != nil {
		// обработать ошибки
		if errors.Is(err, myerrors.ErrAlreadyExists) {
			us.logger.InfoContext(ctx, "user with the same email already exists", "input data", map[string]string{"email": userreq.Email})
			return nil, status.Error(codes.AlreadyExists, "user with the same email already exists")
		}
		us.logger.ErrorContext(logger.ErrorCtx(ctx, err), "unexpected error", "error", err.Error())
		return nil, status.Error(codes.Internal, "Internal Error")
	}
	return &user.LogRegResponse{AccessToken: access, RefreshToken: refresh}, nil
}

func (us *UserService) Login(ctx context.Context, req *user.LoginUserRequest) (*user.LogRegResponse, error) {
	// а насколько важно валидировать данные для входа?
	userreq := models.LoginUserReq{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}
	v := validator.New()
	userreq.Validate(v)
	if !v.Valid() {
		us.logger.InfoContext(ctx, "validation failed", "input data", map[string]string{"email": userreq.Email})
		return logRegBadRequestResponse(v)
	}
	access, refresh, err := us.app.Login(ctx, userreq.Email, userreq.Password)
	if err != nil {
		if errors.Is(err, app.ErrUserNotFound) || errors.Is(err, app.ErrInvalidCredentials) {
			us.logger.InfoContext(ctx, "wrong email or password", "input data", map[string]string{"email": userreq.Email})
			return nil, status.Error(codes.Unauthenticated, "wrong email or password")
		}
		us.logger.ErrorContext(logger.ErrorCtx(ctx, err), "unexpected error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &user.LogRegResponse{AccessToken: access, RefreshToken: refresh}, nil
}

// этот запрос поступает из личного кабинета, поэтому необходимо сверить айди отправителя и айди запрашиваемого аккаунта
func (us *UserService) SendEmailConfirmation(ctx context.Context, req *user.UserID) (*user.Empty, error) {
	id := req.GetId()
	if id == "" {
		us.logger.InfoContext(ctx, "validation failed")
		return nil, status.Error(codes.InvalidArgument, "user id can not be empty")
	}
	//

	// ТУТ ОСТАНОВИЛСЯ С ЛОГЕРОМ

	//
	// мб стоит в горутине это отправлять
	err := us.app.RequestEmailConfirmation(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, app.ErrNoRUID):
			us.logger.ErrorContext(logger.ErrorCtx(ctx, err), "user is not authenticated")
			return nil, status.Error(codes.Unauthenticated, "user is not authenticated")
		case errors.Is(err, app.ErrRUIDneID):
			us.logger.ErrorContext(logger.ErrorCtx(ctx, err), "only the user can request email confirmation")
			return nil, status.Error(codes.PermissionDenied, "only the user can request email confirmation")
		case errors.Is(err, app.ErrUserNotFound):
			us.logger.ErrorContext(logger.ErrorCtx(ctx, err), "no user found")
			return nil, status.Error(codes.NotFound, "no user found")
		case errors.Is(err, app.ErrEmailAlreadyConfirmed):
			us.logger.ErrorContext(logger.ErrorCtx(ctx, err), "email already confirmed")
			return nil, status.Error(codes.AlreadyExists, "email already confirmed")
		case errors.Is(err, app.ErrMsgAlreadySent):
			us.logger.ErrorContext(logger.ErrorCtx(ctx, err), "confirmation letter has already been sent")
			return nil, status.Error(codes.FailedPrecondition, "confirmation letter has already been sent, check your email")
		default:
			us.logger.ErrorContext(logger.ErrorCtx(ctx, err), "unexpected error", "error", err.Error())
			return nil, status.Error(codes.Internal, "something went wrong")
		}
	}
	return &user.Empty{}, nil
}

func (us *UserService) ConfirmEmail(ctx context.Context, req *user.ConfirmEmailRequest) (*user.Empty, error) {
	userID := req.GetUserID()
	if userID == "" {
		us.logger.InfoContext(ctx, "user id can not be empty")
		return nil, status.Error(codes.InvalidArgument, "user id can not be empty")
	}
	mailToken := req.GetMailToken()
	if mailToken == "" {
		us.logger.InfoContext(ctx, "mail token can not be empty")
		return nil, status.Error(codes.InvalidArgument, "mail token can not be empty")
	}
	err := us.app.ConfirmEmail(ctx, userID, mailToken)
	if err != nil {
		switch {
		case errors.Is(err, app.ErrUserNotFound):
			us.logger.InfoContext(logger.ErrorCtx(ctx, err), "user not found")
			return nil, status.Error(codes.NotFound, "user not found")
		case errors.Is(err, app.ErrWrongMailToken):
			us.logger.InfoContext(logger.ErrorCtx(ctx, err), "provided token has been expired or does not exist")
			return nil, status.Error(codes.FailedPrecondition, "provided token has been expired or does not exist")
		default:
			us.logger.ErrorContext(logger.ErrorCtx(ctx, err), "unexpected error", "error", err.Error())
			return nil, status.Error(codes.Internal, "something goes wrong")
		}
	}
	return &user.Empty{}, nil
}

// валидация рефреш токена будет проведена и на фронте, поэтому вызов этого метода
// можно прогнать через любой интерцептор
func (us *UserService) GetNewAccessToken(ctx context.Context, req *user.Token) (*user.Token, error) {
	refresh := req.GetToken()
	if refresh == "" {
		us.logger.InfoContext(ctx, "refresh token can not be empty")
		return nil, status.Error(codes.InvalidArgument, "refresh token can not be empty")
	}
	u, err := us.app.ParseJWTToken(refresh)
	if err != nil {
		us.logger.InfoContext(ctx, "token parsing", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, "provided token badly structured")
	}
	access, err := us.app.CreateJWTToken(u.ID, u.Name, u.IsModer, u.IsAdmin, u.IsCore)
	if err != nil {
		us.logger.ErrorContext(ctx, "token creating", "error", err.Error())
		return nil, status.Error(codes.Internal, "token creating failed")
	}
	return &user.Token{Token: access}, nil
}

// опционально защитить проверкой, чтобы только мои сервисы могли запрашивать
// но можно всё общение защитить mTLS
func (us *UserService) GetRSAPublicKey(ctx context.Context, req *user.Empty) (*user.RSAPublicKey, error) {
	pub, err := us.app.GetRSAPublicKey()
	if err != nil {
		us.logger.ErrorContext(ctx, "PKIX generating failed", "error", err.Error())
		return nil, status.Error(codes.Internal, "PKIX generating failed")
	}
	return &user.RSAPublicKey{
		Kty: "RSA",
		Use: "sig",
		Kid: "1",
		Alg: "RS384",
		Key: pub,
	}, nil
}
