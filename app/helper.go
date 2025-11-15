package app

import (
	"context"

	"github.com/glekoz/online-shop_user/shared/logger"
	"github.com/glekoz/online-shop_user/shared/myerrors"
)

// в интерцепторе уже делается так, чтобы айди не был пустым
func getRUID(ctx context.Context) (string, error) {
	ld, ok := ctx.Value(logger.LogDataKey).(logger.LogData)
	if !ok || ld.UserID == "" {
		return "", myerrors.ErrInvalidCredentials
	}
	return ld.UserID, nil
}
