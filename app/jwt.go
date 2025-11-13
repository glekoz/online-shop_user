package app

import (
	"time"

	"github.com/glekoz/online-shop_user/shared/models"
	"github.com/glekoz/online-shop_user/shared/myerrors"
	"github.com/golang-jwt/jwt/v5"
)

// type data struct {
// 	ID      string
// 	Name    string
// 	IsModer bool
// 	IsAdmin bool
// 	IsCore  bool
// }

// время жизни токена вынести в конфиг 24*time.Hour
func (a *App) CreateJWTToken(userID, name string, isModer, isAdmin, isCore bool) (string, error) {
	return a.createToken(userID, name, isModer, isAdmin, isCore, 15*time.Minute)
}

// время жизни токена вынести в конфиг 15*time.Minute
func (a *App) CreateRefreshToken(userID, name string, isModer, isAdmin, isCore bool) (string, error) {
	return a.createToken(userID, name, isModer, isAdmin, isCore, 24*time.Hour)
}

// возможно, токен будет парситься в http middleware
func (a *App) ParseJWTToken(tokenString string) (models.UserToken, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(token *jwt.Token) (any, error) {
		return a.publicKey, nil
	})
	if err != nil {
		a.logger.Error("jwt.ParseWithClaims")
		return models.UserToken{}, err
	}
	if !token.Valid {
		a.logger.Error("!token.Valid")
		return models.UserToken{}, myerrors.ErrInvalidCredentials
	}
	claims, ok := token.Claims.(*jwt.MapClaims)
	if !ok {
		a.logger.Error("token.Claims.(*jwt.MapClaims)")
		return models.UserToken{}, myerrors.ErrInvalidCredentials
	}
	data, ok := (*claims)["data"].(map[string]any)
	if !ok {
		a.logger.Error("(*claims)[data].(map[string]any)")
		return models.UserToken{}, myerrors.ErrInvalidCredentials
	}
	user, err := models.MapToToken(data)
	if err != nil {
		a.logger.Error(err.Error())
		return models.UserToken{}, myerrors.ErrInvalidCredentials
	}
	return user, nil
}

func (a *App) createToken(userID, name string, isModer, isAdmin, isCore bool, duration time.Duration) (string, error) {
	data := map[string]any{
		"ID":      userID,
		"Name":    name,
		"IsModer": isModer,
		"IsAdmin": isAdmin,
		"IsCore":  isCore,
	}

	claims := &jwt.MapClaims{
		"iss":  "online-shop_user",
		"exp":  jwt.NewNumericDate(time.Now().Add(duration)),
		"data": data,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS384, claims)
	signedToken, err := token.SignedString(a.privateKey)
	if err != nil {
		return "", err
	}
	return signedToken, nil
}
