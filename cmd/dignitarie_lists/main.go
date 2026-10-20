package main

import (
	"context"
	"fmt"

	_ "awesomeProject/docs" // сгенерированная swagger-документация
	"awesomeProject/internal/app/config"
	"awesomeProject/internal/app/dsn"
	"awesomeProject/internal/app/handler"
	"awesomeProject/internal/app/redis"
	"awesomeProject/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Веб-сервис «Списки сановников»
// @version         1.0
// @description     Определение года события по упоминанию сановников. Услуги — списки визирей, консулов и воевод с датами пребывания в должности. Чтение доступно гостю, методы создателя требуют токена: получите его в POST /users/login и нажмите Authorize, вставив значение «Bearer <токен>».

// @host      localhost:8080
// @BasePath  /api
// @schemes   http

// @securityDefinitions.apikey  ApiKeyAuth
// @in                          header
// @name                        Authorization
// @description                 Токен JWT в формате: Bearer <токен>
func main() {
	logrus.Info("Application start!")

	ctx := context.Background()
	cfg := config.NewConfig()

	repo, err := repository.New(dsn.FromEnv(), cfg)
	if err != nil {
		logrus.Fatalf("ошибка инициализации репозитория: %v", err)
	}

	redisClient, err := redis.New(ctx, cfg.Redis)
	if err != nil {
		logrus.Fatalf("ошибка подключения к Redis: %v", err)
	}
	defer redisClient.Close()

	router := gin.Default()
	router.MaxMultipartMemory = 8 << 20
	handler.NewHandler(repo, redisClient, cfg.JWT).RegisterHandler(router)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	if err := router.Run(fmt.Sprintf("%s:%s", cfg.ServiceHost, cfg.ServicePort)); err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("Application terminated!")
}
