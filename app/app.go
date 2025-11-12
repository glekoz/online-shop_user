package app

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"

	"github.com/glekoz/online-shop_user/shared/models"
	"github.com/glekoz/online-shop_user/shared/myerrors"
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

type MailAPI interface {
	SendMessage(recipient, link string) (string, error)
}

type CacheAPI interface {
	Add(userID, token string) error
	Get(userID string) (string, bool)
	Delete(userID string)
}

type App struct {
	Repo      RepoAPI
	Mail      MailAPI
	MailTable CacheAPI
	logger    *slog.Logger

	frontAddr string
	secretKey []byte
}

func New(repo RepoAPI, mail MailAPI, mt CacheAPI, log *slog.Logger, frontAddr string, secretKey []byte) *App {
	return &App{
		Repo:      repo,
		Mail:      mail,
		MailTable: mt,
		logger:    log,

		frontAddr: frontAddr,
		secretKey: secretKey,
	}
}

// для токена возвращается айди и имя, а остальное - false
// не возвращается, а используется
func (a *App) Register(ctx context.Context, name, email, barePassword string) (access string, refresh string, err error) {
	id, err := uuid.NewV7()
	if err != nil {
		a.logger.Error("uuid creating failed", slog.String("email", email))
		return "", "", err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(barePassword), bcrypt.DefaultCost)
	if err != nil {
		a.logger.Error("password hashing failed", slog.String("email", email))
		return "", "", err
	}

	access, err = a.CreateJWTToken(id.String(), name, false, false, false)
	if err != nil {
		a.logger.Error("access jwt token creating failed", slog.String("email", email))
		return "", "", err
	}

	refresh, err = a.CreateRefreshToken(id.String(), name, false, false, false)
	if err != nil {
		a.logger.Error("refresh jwt token creating failed", slog.String("email", email))
		return "", "", err
	}

	err = a.Repo.CreateUser(ctx, id.String(), name, email, string(hashedPassword))
	if err != nil {
		return "", "", err
	}

	a.sendEmailConfirmation(id.String(), email)

	a.logger.Info("user registered, mail sent")
	return access, refresh, nil
}

// мб стоит в горутине это отправлять
// если в горутине, то ошибку можно и игнорировать
func (a *App) sendEmailConfirmation(userID, email string) error {
	mailtoken := rand.Text()
	err := a.MailTable.Add(userID, mailtoken)
	if err != nil {
		a.logger.Error("adding to MailTable failed", slog.String("user ID", userID))
		return err
	}
	// возможно, стоит откатить запись в БД в случае ошибки - зачем? не надо
	_, err = a.Mail.SendMessage(email, fmt.Sprintf("%s/confirm/%s/%s", a.frontAddr, userID, mailtoken))
	if err != nil {
		a.logger.Error("mailing failed", slog.String("email", email), slog.String("error", err.Error()))
		return err
	}
	return nil
}

// этот запрос поступает из личного кабинета, поэтому необходимо сверить айди отправителя и айди запрашиваемого аккаунта
func (a *App) RequestEmailConfirmation(ctx context.Context, userID string) error {
	RUID, err := getRUID(ctx)
	if err != nil {
		return err
	}
	if RUID != userID {
		return myerrors.ErrForbidden
	}
	// проверить, не подтверждена ли уже почта
	user, err := a.Repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.IsEmailConfirmed {
		return myerrors.ErrAlreadyExists
	}

	_, ok := a.MailTable.Get(userID)
	if ok {
		// чтобы не было возможности израскодовать запас писем
		return myerrors.ErrMailSent
	}
	err = a.sendEmailConfirmation(userID, user.Email)
	if err != nil {
		return err
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

// ссылка на этот метод будет в самом письме - тут тоже надо проверять таблицу
func (a *App) ConfirmEmail(ctx context.Context, userID, mailtoken string) error {
	mtoken, ok := a.MailTable.Get(userID)
	if !ok || mtoken != mailtoken {
		return myerrors.ErrNoMailToken
	}

	err := a.Repo.ConfirmEmail(ctx, userID)
	if err != nil {
		return err
	}

	a.MailTable.Delete(userID)
	return nil
}

func (a *App) Login(ctx context.Context, email, barePassword string) (access string, refresh string, err error) {
	user, err := a.Repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", "", err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(barePassword))
	if err != nil {
		return "", "", myerrors.ErrInvalidCredentials
	}
	access, err = a.CreateJWTToken(user.ID, user.Name, user.IsModer, user.IsAdmin, user.IsCore)
	if err != nil {
		a.logger.Error("access jwt token creating failed", slog.String("email", email))
		return "", "", err
	}
	refresh, err = a.CreateRefreshToken(user.ID, user.Name, user.IsModer, user.IsAdmin, user.IsCore)
	if err != nil {
		a.logger.Error("refresh jwt token creating failed", slog.String("email", email))
		return "", "", err
	}
	return access, refresh, nil
}

func (a *App) IssueAccessFromRefresh(ctx context.Context, refresh string) (string, error) {
	user, err := a.ParseJWTToken(refresh)
	if err != nil {
		return "", err
	}
	access, err := a.CreateJWTToken(user.ID, user.Name, user.IsModer, user.IsAdmin, user.IsCore)
	if err != nil {
		a.logger.Error("access jwt token creating failed")
		return "", err
	}
	return access, nil
}

// в будущем можно добавить отправку письма
// в интерсепторе валидирую токен и кладу requestUserID в контекст
// можно ли вынести эту проверку в интерсептор? -- deprecated

// токен валидируется в http middleware
// до сервиса додходит только userID из токена
func (a *App) PromoteModer(ctx context.Context, userID string) error {
	RUID, err := getRUID(ctx)
	if err != nil {
		return err
	}
	_, err = a.Repo.GetAdmin(ctx, RUID)
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
	RUID, err := getRUID(ctx)
	if err != nil {
		return err
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
	RUID, err := getRUID(ctx)
	if err != nil {
		return err
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
	// RUID, ok := ctx.Value(vars.ContextKeyRequestUserID).(string)
	// if !ok || RUID == "" {
	// 	return models.User{}, myerrors.ErrInvalidCredentials
	// }
	user, err := a.Repo.GetUserByID(ctx, userID)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

// когда не админ, редирект на свою страницу
func (a *App) GetUsersByEmail(ctx context.Context, email string) ([]models.UserInfo, error) {
	RUID, err := getRUID(ctx)
	if err != nil {
		return nil, err
	}
	_, err = a.Repo.GetAdmin(ctx, RUID)
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
	// RUID, ok := ctx.Value(vars.ContextKeyRequestUserID).(string)
	// if !ok || RUID == "" {
	// 	return myerrors.ErrInvalidCredentials
	// }
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
	RUID, err := getRUID(ctx)
	if err != nil {
		return err
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
