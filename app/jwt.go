package app

import (
	"time"

	"github.com/glekoz/online-shop_user/shared/myerrors"
	"github.com/golang-jwt/jwt/v5"
)

type data struct {
	ID      string
	Name    string
	IsModer bool
	IsAdmin bool
	IsCore  bool
}

// время жизни токена вынести в конфиг
func (a *App) CreateJWTToken(userID, name string, isModer, isAdmin, isCore bool) (string, error) {
	claims := &jwt.MapClaims{
		"iss": "online-shop_user",
		"exp": jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		"data": data{
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

// возможно, токен будет парситься в http middleware
func (a *App) ParseJWTToken(tokenString string) (data, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return a.secretKey, nil
	})
	if err != nil {
		return data{}, err
	}
	claims, ok := token.Claims.(*jwt.MapClaims)
	if !ok || !token.Valid {
		return data{}, myerrors.ErrInvalidToken
	}
	d, ok := (*claims)["data"].(data)
	if !ok {
		return data{}, myerrors.ErrInvalidToken
	}
	return d, nil
}
