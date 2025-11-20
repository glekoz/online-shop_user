package handler

import (
	"context"
	"log/slog"

	"github.com/glekoz/online-shop_proto/user"
	"github.com/glekoz/online-shop_user/shared/models"
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
		return nil, us.handleError(ctx, err, "input data", map[string]string{"email": userreq.Email})
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
		return nil, us.handleError(ctx, err, "input data", map[string]string{"email": userreq.Email})
	}
	return &user.LogRegResponse{AccessToken: access, RefreshToken: refresh}, nil
}

// этот запрос поступает из личного кабинета, поэтому необходимо сверить айди отправителя и айди запрашиваемого аккаунта
func (us *UserService) SendEmailConfirmation(ctx context.Context, req *user.UserID) (*user.Empty, error) {
	id := req.GetId()
	if id == "" {
		us.logger.InfoContext(ctx, "validation failed")
		return nil, badRequestResponse("validation", map[string]string{"id": "must be provided"})
	}

	// мб стоит в горутине это отправлять
	err := us.app.RequestEmailConfirmation(ctx, id)
	if err != nil {
		return nil, us.handleError(ctx, err)
	}
	return &user.Empty{}, nil
}

func (us *UserService) ConfirmEmail(ctx context.Context, req *user.ConfirmEmailRequest) (*user.Empty, error) {
	userID := req.GetUserID()
	if userID == "" {
		us.logger.InfoContext(ctx, "validation failed", slog.String("user id", "must be provided"))
		return nil, badRequestResponse("validation", map[string]string{"user id": "must be provided"})
	}
	mailToken := req.GetMailToken()
	if mailToken == "" {
		us.logger.InfoContext(ctx, "validation failed", slog.String("mail token", "must be provided"))
		return nil, badRequestResponse("validation", map[string]string{"mail token": "must be provided"})
	}
	err := us.app.ConfirmEmail(ctx, userID, mailToken)
	if err != nil {
		return nil, us.handleError(ctx, err)
	}
	return &user.Empty{}, nil
}

// валидация рефреш токена будет проведена и на фронте, поэтому вызов этого метода
// можно прогнать через любой интерцептор
func (us *UserService) GetNewAccessToken(ctx context.Context, req *user.Token) (*user.Token, error) {
	refresh := req.GetToken()
	if refresh == "" {
		us.logger.InfoContext(ctx, "validation failed", slog.String("refresh token", "must be provided"))
		return nil, badRequestResponse("validation", map[string]string{"refresh token": "must be provided"})
	}
	u, err := us.app.ParseJWTToken(refresh)
	if err != nil {
		us.logger.InfoContext(ctx, "token parsing", "error", err.Error())
		return nil, badRequestResponse("parsing", map[string]string{"refresh token": "badly structured"})
	}
	access, err := us.app.CreateJWTToken(u.ID, u.Name, u.IsModer, u.IsAdmin, u.IsCore)
	if err != nil {
		return nil, us.handleError(ctx, err)
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
