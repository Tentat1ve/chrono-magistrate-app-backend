package api

import (
	"log"
	"net/http"

	"awesomeProject/internal/app/handler"
	"awesomeProject/internal/app/repository"

	"github.com/gin-gonic/gin"
)

func StartServer(minioURL string) {
	log.Println("Server start up")

	repo := repository.NewRepository(minioURL)
	h := handler.NewHandler(repo)

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./static")

	h.RegisterHandler(r)

	// несуществующие адреса перенаправляются на плитку вместо ответа 404
	r.NoRoute(func(ctx *gin.Context) {
		ctx.Redirect(http.StatusSeeOther, "/dignitarie_lists/catalog")
	})

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
	log.Println("Server down")
}
