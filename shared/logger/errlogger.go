package logger

import (
	"context"
	"errors"
)

type ErrorLogData struct {
	LD  LogData
	Err error
}

func (e *ErrorLogData) Error() string {
	return e.Err.Error()
}

// WrapError нужен, когда вместе с ошибкой нужно передать дополнительные поля
// если никакой дополнительной информации нет, то и оборачивать нечем
func WrapError(ctx context.Context, err error) error {
	ld := LogData{}
	if ldt, ok := ctx.Value(LogDataKey).(LogData); ok {
		ld = ldt
	}
	return &ErrorLogData{LD: ld, Err: err}
}

func ErrorCtx(ctx context.Context, err error) context.Context {
	var errld *ErrorLogData
	if errors.As(err, &errld) {
		return context.WithValue(ctx, LogDataKey, errld.LD)
	}
	return ctx
}
