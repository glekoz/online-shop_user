package main

import (
	"crypto/rand"
	"crypto/rsa"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/glekoz/online-shop_user/app"
	"github.com/glekoz/online-shop_user/cache"
	"github.com/glekoz/online-shop_user/handler"
	"github.com/glekoz/online-shop_user/mail"
	"github.com/glekoz/online-shop_user/repository"
)

func main() {
	privateKey, err := genKey()
	if err != nil {
		panic(err)
	}
	repo, err := repository.New("postgres://postgres:postgres@localhost:5432/training?sslmode=disable")
	if err != nil {
		panic(err)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				a.Value = slog.TimeValue(a.Value.Time().Truncate(time.Minute))
			}
			return a
		},
		Level: slog.LevelInfo,
	}))
	mail := mail.New()
	c, err := cache.New(3600)
	if err != nil {
		log.Fatal("cache issue")
	}
	app := app.New(repo, mail, c, logger, "frontAddr", privateKey, &privateKey.PublicKey)
	server := handler.NewServer(app, logger)
	logger.Info("starting grpc server...")
	server.RunServer(8080)
}

// заменить на чтение из конфигурации
func genKey() (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}
