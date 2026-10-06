package main

import (
	"awesomeProject/internal/app/config"
	"awesomeProject/internal/app/dsn"
	"awesomeProject/internal/app/handler"
	"awesomeProject/internal/app/repository"
	"awesomeProject/internal/pkg"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.Info("Application start!")

	cfg := config.NewConfig()

	repo, err := repository.New(dsn.FromEnv())
	if err != nil {
		logrus.Fatalf("ошибка подключения к БД: %v", err)
	}

	h := handler.NewHandler(repo, cfg.MinioPublicURL)
	app := pkg.NewApp(cfg, gin.Default(), h)
	app.RunApp()

	logrus.Info("Application terminated!")
}
