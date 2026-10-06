package main

import (
	"log"
	"os"

	"awesomeProject/internal/api"
)

func main() {
	log.Println("Application start!")

	// Публичный адрес бакета Minio, из которого браузер загружает изображения и видео
	minioURL := os.Getenv("MINIO_PUBLIC_URL")
	if minioURL == "" {
		minioURL = "http://localhost:9000/dignitaries"
	}

	api.StartServer(minioURL)
	log.Println("Application terminated!")
}
