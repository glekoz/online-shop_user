package app

import (
	"context"

	"github.com/glekoz/online-shop_user/shared/myerrors"
	"github.com/glekoz/online-shop_user/shared/vars"
)

// в интерцепторе уже делается так, чтобы айди не был пустым
func getRUID(ctx context.Context) (string, error) {
	RUID, ok := ctx.Value(vars.ContextKeyRequestUserID).(string)
	if !ok || RUID == "" {
		return "", myerrors.ErrInvalidCredentials
	}
	return RUID, nil
}
