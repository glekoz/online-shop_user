package app

import (
	"context"
	"errors"

	"github.com/glekoz/online-shop_user/shared/models"
	"github.com/glekoz/online-shop_user/shared/myerrors"
	"github.com/glekoz/online-shop_user/shared/vars"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type RepoAPI interface {
	CreateUser(ctx context.Context, id, name, email, hashedPassword string) error
	PromoteModer(ctx context.Context, id string) error
	PromoteAdmin(ctx context.Context, id string) error
	PromoteCoreAdmin(ctx context.Context, id string) error
	GetUserByID(ctx context.Context, id string) (models.User, error)
	GetUserByEmail(ctx context.Context, email string) (models.UserToken, error)
	GetUsersByEmail(ctx context.Context, email string) ([]models.UserInfo, error)
	GetModer(ctx context.Context, id string) (string, error)
	GetAdmin(ctx context.Context, id string) (models.Admin, error)
	ConfirmEmail(ctx context.Context, id string) error
	ChangeName(ctx context.Context, id, newName string) error
	ChangePassword(ctx context.Context, id, newHashedPassword string) error
	ChangeEmail(ctx context.Context, id, newEmail string) error
	DeleteUser(ctx context.Context, id string) error
	DeleteModer(ctx context.Context, id string) error
	DeleteAdmin(ctx context.Context, id string) error
}

type App struct {
	Repo RepoAPI
	// emailSender EmailSenderAPI
	secretKey []byte
}

func New(repo RepoAPI, secretKey []byte) *App {
	return &App{
		Repo:      repo,
		secretKey: secretKey,
	}
}

// для токена возвращается айди и имя, а остальное - false
// не возвращается, а используется
func (a *App) RegisterUser(ctx context.Context, name, email, barePassword string) (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(barePassword), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	err = a.Repo.CreateUser(ctx, id.String(), name, email, string(hashedPassword))
	if err != nil {
		return "", err
	}
	token, err := a.CreateJWTToken(id.String(), name, false, false, false)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (a *App) ConfirmEmailRequest(ctx context.Context, userID string) error {
	// проверить, не подтверждена ли уже почта
	user, err := a.Repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.IsEmailConfirmed {
		return myerrors.ErrAlreadyExists
	}
	// генерация токена и сохранение его в кэше вместе с userID
	//
	// отправка письма с ссылкой на подтверждение почты
	// ссылка типа /confirm/<uid>/<token>, где токен хранится в кэше в паре userID : token
	// и при переходе по ссылке посмотреть, валиден ли токен для этого пользователя
	// и если да, то подтвердить почту
	//
	return nil
}

func (a *App) ConfirmEmail(ctx context.Context, userID string) error {
	err := a.Repo.ConfirmEmail(ctx, userID)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) LoginUser(ctx context.Context, email, barePassword string) (string, error) {
	user, err := a.Repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(barePassword))
	if err != nil {
		return "", myerrors.ErrInvalidCredentials
	}
	token, err := a.CreateJWTToken(user.ID, user.Name, user.IsModer, user.IsAdmin, user.IsCore)
	if err != nil {
		return "", err
	}
	return token, nil
}

// в будущем можно добавить отправку письма
// в интерсепторе валидирую токен и кладу requestUserID в контекст
// можно ли вынести эту проверку в интерсептор? -- deprecated

// токен валидируется в http middleware
// до сервиса додходит только userID из токена
func (a *App) PromoteModer(ctx context.Context, userID string) error {
	RUID, ok := ctx.Value(vars.ContextKeyRequestUserID).(string)
	if !ok || RUID == "" {
		return myerrors.ErrInvalidCredentials
	}
	_, err := a.Repo.GetAdmin(ctx, RUID)
	if err != nil {
		if errors.Is(err, myerrors.ErrNotFound) {
			return myerrors.ErrForbidden
		}
		return err
	}
	err = a.Repo.PromoteModer(ctx, userID)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) PromoteAdmin(ctx context.Context, userID string, isCore bool) error {
	RUID, ok := ctx.Value(vars.ContextKeyRequestUserID).(string)
	if !ok || RUID == "" {
		return myerrors.ErrInvalidCredentials
	}
	admin, err := a.Repo.GetAdmin(ctx, RUID)
	if err != nil {
		if errors.Is(err, myerrors.ErrNotFound) {
			return myerrors.ErrForbidden
		}
		return err
	}
	if !admin.IsCore {
		return myerrors.ErrForbidden
	}
	err = a.Repo.PromoteAdmin(ctx, userID)
	if err != nil {
		return err
	}
	return nil
}

// специально разнесен с PromoteAdmin
func (a *App) PromoteCoreAdmin(ctx context.Context, userID string) error {
	RUID, ok := ctx.Value(vars.ContextKeyRequestUserID).(string)
	if !ok || RUID == "" {
		return myerrors.ErrInvalidCredentials
	}
	admin, err := a.Repo.GetAdmin(ctx, RUID)
	if err != nil {
		if errors.Is(err, myerrors.ErrNotFound) {
			return myerrors.ErrForbidden
		}
		return err
	}
	if !admin.IsCore {
		return myerrors.ErrForbidden
	}
	err = a.Repo.PromoteCoreAdmin(ctx, userID)
	if err != nil {
		return err
	}
	return nil
}

// открыть страницу пользователя может только авторизованный пользователь
func (a *App) GetUserByID(ctx context.Context, userID string) (models.User, error) {
	RUID, ok := ctx.Value(vars.ContextKeyRequestUserID).(string)
	if !ok || RUID == "" {
		return models.User{}, myerrors.ErrInvalidCredentials
	}
	user, err := a.Repo.GetUserByID(ctx, userID)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

// когда не админ, редирект на свою страницу
func (a *App) GetUsersByEmail(ctx context.Context, email string) ([]models.UserInfo, error) {
	RUID, ok := ctx.Value(vars.ContextKeyRequestUserID).(string)
	if !ok || RUID == "" {
		return nil, myerrors.ErrInvalidCredentials
	}
	_, err := a.Repo.GetAdmin(ctx, RUID)
	if err != nil {
		if errors.Is(err, myerrors.ErrNotFound) {
			return nil, myerrors.ErrForbidden
		}
		return nil, err
	}
	users, err := a.Repo.GetUsersByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (a *App) ChangePasswordRequest(ctx context.Context) error {
	RUID, ok := ctx.Value(vars.ContextKeyRequestUserID).(string)
	if !ok || RUID == "" {
		return myerrors.ErrInvalidCredentials
	}
	// генерация токена и сохранение его в кэше вместе с userID
	//
	// отправка письма с ссылкой на смену пароля
	// ссылка типа /reset_password/<uid>/<token>, где токен хранится в кэше в паре userID : token
	// и при переходе по ссылке посмотреть, валиден ли токен для этого пользователя
	// и если да, то разрешить смену пароля
	//
	return nil
}

func (a *App) ResetPasswordRequest(ctx context.Context, email string) error {
	_, err := a.Repo.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	// генерация токена и сохранение его в кэше вместе с userID
	//
	// отправка письма с ссылкой на сброс пароля
	// ссылка типа /reset_password/<uid>/<token>, где токен хранится в кэше в паре userID : token
	// и при переходе по ссылке посмотреть, валиден ли токен для этого пользователя
	// и если да, то разрешить сброс пароля
	//
	return nil
}

// в любом случае приходит ссылка на почту,
// поэтому нет смысла запрашивать старый пароль
func (a *App) ChangePassword(ctx context.Context, newBarePassword string) error {
	RUID, ok := ctx.Value(vars.ContextKeyRequestUserID).(string)
	if !ok || RUID == "" {
		return myerrors.ErrInvalidCredentials
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newBarePassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	err = a.Repo.ChangePassword(ctx, RUID, string(hashedPassword))
	if err != nil {
		return err
	}
	return nil
}

// ----------------------------------------------------------------------
// ДЛЯ ДРУГИХ СЕРВИСОВ
// ----------------------------------------------------------------------

func (a *App) IsAdmin(ctx context.Context, userID string) (bool, error) {
	_, err := a.Repo.GetAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, myerrors.ErrNotFound) {
			return false, myerrors.ErrForbidden
		}
		return false, err
	}
	return true, nil
}

func (a *App) IsModer(ctx context.Context, userID string) (bool, error) {
	_, err := a.Repo.GetModer(ctx, userID)
	if err != nil {
		if errors.Is(err, myerrors.ErrNotFound) {
			return false, myerrors.ErrForbidden
		}
		return false, err
	}
	return true, nil
}
