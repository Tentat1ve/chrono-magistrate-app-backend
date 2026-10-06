package api

import (
	"fmt"
	"html/template"
	"log"

	"awesomeProject/internal/app/handler"
	"awesomeProject/internal/app/repository"

	"github.com/gin-gonic/gin"
)

// formatYear выводит год с учётом эры: -63 → «63 г. до н.э.», 1565 → «1565 г.»
func formatYear(year int) string {
	if year < 0 {
		return fmt.Sprintf("%d г. до н.э.", -year)
	}
	return fmt.Sprintf("%d г.", year)
}

func StartServer(minioURL string) {
	log.Println("Server start up")

	repo := repository.NewRepository(minioURL)
	h := handler.NewHandler(repo)

	r := gin.Default()
	r.SetFuncMap(template.FuncMap{"formatYear": formatYear})
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./static")

	r.GET("/dignitaries", h.GetDignitaries)
	r.GET("/dignitary_feed", h.GetDignitaryFeed)
	r.GET("/dignitary_feed/:id", h.GetDignitaryFeed)
	r.GET("/dignitary_draft", h.GetDignitaryDraft)

	r.Run(":8080")
	log.Println("Server down")
}
