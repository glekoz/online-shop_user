package app

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"

	"github.com/glekoz/online-shop_user/mail"
	"github.com/glekoz/online-shop_user/repository"
	"github.com/glekoz/online-shop_user/shared/logger"
	"github.com/glekoz/online-shop_user/shared/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type RepoAPI interface {
	CreateUser(ctx context.Context, id, name, email, hashedPassword string) error
	PromoteModer(ctx context.Context, id string) error
	PromoteAdmin(ctx context.Context, id string) error
	PromoteCoreAdmin(ctx context.Context, id string) error
	GetUserByID(ctx context.Context, id string) (models.User, error)
	GetUserByEmail(ctx context.Context, email string) (models.UserTokenWithPassword, error)
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
	SendEmailConfirmationMessage(userID, email string, mailtoken, link string) (string, error)
	CheckToken(userID, token string) bool
}
type CacheAPI interface {
	Add(userID, token string) error
	Get(userID string) (string, bool)
	Delete(userID string)
}

type App struct {
	Repo RepoAPI
	Mail MailAPI
	// MailTable CacheAPI
	logger *slog.Logger

	frontAddr  string
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func New(repo RepoAPI, mail MailAPI, mt CacheAPI, log *slog.Logger, frontAddr string, privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) *App {
	return &App{
		Repo: repo,
		Mail: mail,
		// MailTable: mt,
		logger: log,

		frontAddr:  frontAddr,
		privateKey: privateKey,
		publicKey:  publicKey,
	}
}

// для токена возвращается айди и имя, а остальное - false
// не возвращается, а используется
func (a *App) Register(ctx context.Context, name, email, barePassword string) (access string, refresh string, err error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", "", err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(barePassword), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}

	err = a.Repo.CreateUser(ctx, id.String(), name, email, string(hashedPassword))
	if err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			return "", "", ErrUserAlreadyExists
		}
		return "", "", err
	}

	access, err = a.CreateJWTToken(id.String(), name, false, false, false)
	if err != nil {
		return "", "", err
	}

	refresh, err = a.CreateRefreshToken(id.String(), name, false, false, false)
	if err != nil {
		return "", "", err
	}

	msgID, err := a.sendEmailConfirmation(id.String(), email)
	if err != nil {
		if errors.Is(err, mail.ErrMsgAlreadySent) {
			a.logger.InfoContext(ctx, "email message has already been sent", "input data", map[string]string{"email": email})
		} else {
			a.logger.ErrorContext(ctx, "mail malfunction", "data", map[string]string{"email": email})
		}
	}
	a.logger.InfoContext(ctx, "email sent", "data", map[string]string{"email": email, "msgID": msgID})

	return access, refresh, nil
}

// мб стоит в горутине это отправлять
// если в горутине, то ошибку можно и игнорировать
func (a *App) sendEmailConfirmation(userID, email string) (string, error) {
	mailtoken := rand.Text()
	link := fmt.Sprintf("%s/confirm/%s/%s", a.frontAddr, userID, mailtoken)
	msgID, err := a.Mail.SendEmailConfirmationMessage(userID, email, mailtoken, link)
	if err != nil {
		return "", err
	}
	return msgID, nil
}

// этот запрос поступает из личного кабинета, поэтому необходимо сверить айди отправителя и айди запрашиваемого аккаунта
func (a *App) RequestEmailConfirmation(ctx context.Context, userID string) error {
	RUID, err := getRUID(ctx)
	if err != nil {
		return ErrNoRUID
	}
	if RUID != userID {
		ctx = logger.WithDetails(ctx, "id", userID)
		return logger.WrapError(ctx, ErrRUIDneID)
	}
	// проверить, не подтверждена ли уже почта
	user, err := a.Repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	ctx = logger.WithDetails(ctx, "email", user.Email)
	if user.IsEmailConfirmed {
		return logger.WrapError(ctx, ErrEmailAlreadyConfirmed)
	}

	msgID, err := a.sendEmailConfirmation(userID, user.Email)
	if err != nil {
		if errors.Is(err, mail.ErrMsgAlreadySent) {
			return logger.WrapError(ctx, ErrMsgAlreadySent)
		}
		return logger.WrapError(ctx, err)
	}
	a.logger.InfoContext(ctx, "email sent", "msgID", msgID)

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
	ok := a.Mail.CheckToken(userID, mailtoken)
	if !ok {
		ctx = logger.WithDetails(ctx, "mail token", mailtoken)
		return logger.WrapError(ctx, ErrWrongMailToken)
	}

	err := a.Repo.ConfirmEmail(ctx, userID)
	if err != nil {
		ctx = logger.WithDetails(ctx, "id", userID)
		if errors.Is(err, repository.ErrNotFound) {
			return logger.WrapError(ctx, ErrUserNotFound)
		}
		return logger.WrapError(ctx, err)
	}
	return nil
}

func (a *App) Login(ctx context.Context, email, barePassword string) (access string, refresh string, err error) {
	user, err := a.Repo.GetUserByEmail(ctx, email)
	if err != nil {
		ctx = logger.WithDetails(ctx, "email", email)
		if errors.Is(err, repository.ErrNotFound) {
			return "", "", logger.WrapError(ctx, ErrUserNotFound)
		}
		return "", "", logger.WrapError(ctx, err)
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(barePassword))
	if err != nil {
		return "", "", ErrInvalidCredentials
	}
	access, err = a.CreateJWTToken(user.ID, user.Name, user.IsModer, user.IsAdmin, user.IsCore)
	if err != nil {
		return "", "", err
	}
	refresh, err = a.CreateRefreshToken(user.ID, user.Name, user.IsModer, user.IsAdmin, user.IsCore)
	if err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

func (a *App) IssueAccessFromRefresh(ctx context.Context, refresh string) (string, error) {
	user, err := a.ParseJWTToken(refresh)
	if err != nil {
		// ошибку надо или логировать, или передавать выше, но тут для дебага и то, и другое оставлю
		a.logger.InfoContext(ctx, "parse token", slog.String("error", err.Error()))
		return "", ErrInvalidCredentials
	}
	access, err := a.CreateJWTToken(user.ID, user.Name, user.IsModer, user.IsAdmin, user.IsCore)
	if err != nil {
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
		return ErrNoRUID
	}
	_, err = a.Repo.GetAdmin(ctx, RUID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrForbidden
		}
		return err
	}
	err = a.Repo.PromoteModer(ctx, userID)
	if err != nil {
		ctx = logger.WithDetails(ctx, "id", userID)
		if errors.Is(err, repository.ErrNotFound) {
			return logger.WrapError(ctx, ErrUserNotFound)
		}
		if errors.Is(err, repository.ErrAlreadyExists) {
			return logger.WrapError(ctx, ErrUserAlreadyExists)
		}
		return logger.WrapError(ctx, err)
	}
	return nil
}

func (a *App) PromoteAdmin(ctx context.Context, userID string) error {
	RUID, err := getRUID(ctx)
	if err != nil {
		return ErrNoRUID
	}
	//тут не доделал обработку ошибок с репозитория
	admin, err := a.Repo.GetAdmin(ctx, RUID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrForbidden
		}
		return err
	}
	if !admin.IsCore {
		return ErrForbidden
	}
	//тут не доделал обработку ошибок с репозитория
	err = a.Repo.PromoteAdmin(ctx, userID)
	if err != nil {
		ctx = logger.WithDetails(ctx, "id", userID)
		if errors.Is(err, repository.ErrNotFound) {
			return logger.WrapError(ctx, ErrUserNotFound)
		}
		if errors.Is(err, repository.ErrAlreadyExists) {
			return logger.WrapError(ctx, ErrUserAlreadyExists)
		}
		return logger.WrapError(ctx, err)
	}
	return nil
}

// специально разнесен с PromoteAdmin
func (a *App) PromoteCoreAdmin(ctx context.Context, userID string) error {
	RUID, err := getRUID(ctx)
	if err != nil {
		return ErrNoRUID
	}
	//тут не доделал обработку ошибок с репозитория
	admin, err := a.Repo.GetAdmin(ctx, RUID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrForbidden
		}
		return err
	}
	if !admin.IsCore {
		return ErrForbidden
	}
	//тут не доделал обработку ошибок с репозитория
	err = a.Repo.PromoteCoreAdmin(ctx, userID)
	if err != nil {
		ctx = logger.WithDetails(ctx, "id", userID)
		if errors.Is(err, repository.ErrNotFound) {
			return logger.WrapError(ctx, ErrUserNotFound)
		}
		return logger.WrapError(ctx, err)
	}
	return nil
}

// открыть страницу пользователя может только авторизованный пользователь
func (a *App) GetUserByID(ctx context.Context, userID string) (models.User, error) {
	// RUID, ok := ctx.Value(vars.ContextKeyRequestUserID).(string)
	// if !ok || RUID == "" {
	// 	return models.User{}, myerrors.ErrInvalidCredentials
	// }

	//тут не доделал обработку ошибок с репозитория
	user, err := a.Repo.GetUserByID(ctx, userID)
	if err != nil {
		ctx = logger.WithDetails(ctx, "id", userID)
		if errors.Is(err, repository.ErrNotFound) {
			return models.User{}, logger.WrapError(ctx, ErrUserNotFound)
		}
		return models.User{}, logger.WrapError(ctx, err)
	}
	return user, nil
}

// когда не админ, редирект на свою страницу
func (a *App) GetUsersByEmail(ctx context.Context, email string) ([]models.UserInfo, error) {
	RUID, err := getRUID(ctx)
	if err != nil {
		return nil, ErrNoRUID
	}
	_, err = a.Repo.GetAdmin(ctx, RUID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrForbidden
		}
		return nil, err
	}
	users, err := a.Repo.GetUsersByEmail(ctx, email)
	if err != nil {
		ctx = logger.WithDetails(ctx, "email", email)
		if errors.Is(err, repository.ErrNotFound) {
			return nil, logger.WrapError(ctx, ErrUserNotFound)
		}
		return nil, logger.WrapError(ctx, err)
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
		ctx = logger.WithDetails(ctx, "email", email)
		if errors.Is(err, repository.ErrNotFound) {
			return logger.WrapError(ctx, ErrUserNotFound)
		}
		return logger.WrapError(ctx, err)
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
		return ErrNoRUID
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
	//тут не доделал обработку ошибок с репозитория
	_, err := a.Repo.GetAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return false, ErrForbidden
		}
		return false, err
	}
	return true, nil
}

func (a *App) IsModer(ctx context.Context, userID string) (bool, error) {
	//тут не доделал обработку ошибок с репозитория
	_, err := a.Repo.GetModer(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return false, ErrForbidden
		}
		return false, err
	}
	return true, nil
}

func (a *App) GetRSAPublicKey() ([]byte, error) {
	der, err := x509.MarshalPKIXPublicKey(a.publicKey)
	if err != nil {
		return nil, err
	}
	res := make([]byte, base64.StdEncoding.Strict().EncodedLen(len(der)))
	base64.StdEncoding.Encode(res, der)
	return res, nil
}
