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

// время жизни токена вынести в конфиг 24 * time.Hour
func (a *App) CreateJWTToken(userID, name string, isModer, isAdmin, isCore bool) (string, error) {
	return a.createToken(userID, name, isModer, isAdmin, isCore, 15*time.Minute)
}

func (a *App) CreateRefreshToken(userID, name string, isModer, isAdmin, isCore bool) (string, error) {
	return a.createToken(userID, name, isModer, isAdmin, isCore, 24*time.Hour)
}

// возможно, токен будет парситься в http middleware
func (a *App) ParseJWTToken(tokenString string) (models.UserToken, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(token *jwt.Token) (any, error) {
		return a.secretKey, nil
	})
	if err != nil {
		return models.UserToken{}, err
	}
	claims, ok := token.Claims.(*jwt.MapClaims)
	if !ok || !token.Valid {
		return models.UserToken{}, myerrors.ErrInvalidCredentials
	}
	d, ok := (*claims)["sub"].(models.UserToken)
	if !ok {
		return models.UserToken{}, myerrors.ErrInvalidCredentials
	}
	return d, nil
}

func (a *App) createToken(userID, name string, isModer, isAdmin, isCore bool, duration time.Duration) (string, error) {
	claims := &jwt.MapClaims{
		"iss": "online-shop_user",
		"exp": jwt.NewNumericDate(time.Now().Add(duration)),
		"sub": models.UserToken{
			ID:      userID,
			Name:    name,
			IsModer: isModer,
			IsAdmin: isAdmin,
			IsCore:  isCore,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(a.secretKey)
	if err != nil {
		return "", err
	}
	return signedToken, nil
}
