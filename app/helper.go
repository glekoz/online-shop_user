package app

import (
	"context"

	"github.com/glekoz/online-shop_user/shared/logger"
)

// в интерцепторе уже делается так, чтобы айди не был пустым
// чтобы айди БЫЛ, иначе будет ошибка для тех методов, которые
// требуют аутентификацию
func getRUID(ctx context.Context) (string, error) {
	ld, ok := ctx.Value(logger.LogDataKey).(logger.LogData)
	if !ok || ld.UserID == "" {
		return "", ErrNoRUID
	}
	return ld.UserID, nil
}
