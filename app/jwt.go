package app

import (
	"errors"
	"time"

	"github.com/glekoz/online-shop_user/shared/models"
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
// любая ошибка говорит о том, что что-то не так с токеном
func (a *App) ParseJWTToken(tokenString string) (models.UserToken, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(token *jwt.Token) (any, error) {
		return a.publicKey, nil
	})
	if err != nil {
		return models.UserToken{}, err
	}
	if !token.Valid {
		return models.UserToken{}, errors.New("!token.Valid")
	}
	claims, ok := token.Claims.(*jwt.MapClaims)
	if !ok {
		return models.UserToken{}, errors.New("token.Claims.(*jwt.MapClaims)")
	}
	data, ok := (*claims)["data"].(map[string]any)
	if !ok {
		return models.UserToken{}, errors.New("(*claims)[data].(map[string]any)")
	}
	user, err := models.MapToToken(data)
	if err != nil {
		return models.UserToken{}, err
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
