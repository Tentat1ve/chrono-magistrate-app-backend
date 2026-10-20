package main

import (
	"fmt"

	"awesomeProject/internal/app/config"
	"awesomeProject/internal/app/dsn"
	"awesomeProject/internal/app/handler"
	"awesomeProject/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.Info("Application start!")

	cfg := config.NewConfig()

	repo, err := repository.New(dsn.FromEnv(), cfg)
	if err != nil {
		logrus.Fatalf("ошибка инициализации репозитория: %v", err)
	}

	router := gin.Default()
	router.MaxMultipartMemory = 8 << 20
	handler.NewHandler(repo).RegisterHandler(router)

	if err := router.Run(fmt.Sprintf("%s:%s", cfg.ServiceHost, cfg.ServicePort)); err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("Application terminated!")
}
