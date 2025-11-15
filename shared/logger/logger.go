package logger

import (
	"context"
	"io"
	"log/slog"
	"time"
)

type customKey int

const LogDataKey = customKey(0)

type MyJSONLogHandler struct {
	handler slog.Handler
}

type LogData struct {
	UserID    string
	IPAddress string
	Method    string
}

func NewMyJSONLogHandler(n slog.Handler) *MyJSONLogHandler {
	return &MyJSONLogHandler{handler: n}
}

func (h *MyJSONLogHandler) Enabled(ctx context.Context, lvl slog.Level) bool {
	return h.handler.Enabled(ctx, lvl)
}

func (h *MyJSONLogHandler) Handle(ctx context.Context, rec slog.Record) error {
	if ld, ok := ctx.Value(LogDataKey).(LogData); ok {
		if ld.UserID != "" {
			rec.Add("user_id", ld.UserID)
		}
		if ld.IPAddress != "" {
			rec.Add("ip", ld.IPAddress)
		}
		if ld.Method != "" {
			rec.Add("method", ld.Method)
		}
	}
	return h.handler.Handle(ctx, rec)
}

func (h *MyJSONLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h.handler.WithAttrs(attrs)
}

func (h *MyJSONLogHandler) WithGroup(name string) slog.Handler {
	return h.handler.WithGroup(name)
}

func WithUserID(ctx context.Context, userID string) context.Context {
	if ld, ok := ctx.Value(LogDataKey).(LogData); ok {
		ld.UserID = userID
		return context.WithValue(ctx, LogDataKey, ld)
	}
	return context.WithValue(ctx, LogDataKey, LogData{UserID: userID})
}

func WithIPAddress(ctx context.Context, ipAddress string) context.Context {
	if ld, ok := ctx.Value(LogDataKey).(LogData); ok {
		ld.IPAddress = ipAddress
		return context.WithValue(ctx, LogDataKey, ld)
	}
	return context.WithValue(ctx, LogDataKey, LogData{IPAddress: ipAddress})
}

func WithMethod(ctx context.Context, method string) context.Context {
	if ld, ok := ctx.Value(LogDataKey).(LogData); ok {
		ld.Method = method
		return context.WithValue(ctx, LogDataKey, ld)
	}
	return context.WithValue(ctx, LogDataKey, LogData{Method: method})
}

func New(w io.Writer, opts *slog.HandlerOptions) *slog.Logger {
	if opts != nil {
		return slog.New(NewMyJSONLogHandler(slog.Handler(slog.NewJSONHandler(w, opts))))
	}
	handler := slog.Handler(slog.NewJSONHandler(w, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				a.Value = slog.TimeValue(a.Value.Time().Truncate(time.Minute))
			}
			return a
		},
		Level: slog.LevelInfo,
	}))
	handler = NewMyJSONLogHandler(handler)
	return slog.New(handler)
}
