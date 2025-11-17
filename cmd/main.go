package main

import (
	"crypto/rand"
	"crypto/rsa"
	"log"
	"os"

	"github.com/glekoz/online-shop_user/app"
	"github.com/glekoz/online-shop_user/cache"
	"github.com/glekoz/online-shop_user/handler"
	"github.com/glekoz/online-shop_user/mail"
	"github.com/glekoz/online-shop_user/repository"
	"github.com/glekoz/online-shop_user/shared/logger"
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
	logger := logger.New(os.Stdout, nil)
	c1, err := cache.New(3600)
	if err != nil {
		log.Fatal("cache issue")
	}
	mail := mail.New(c1)
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
