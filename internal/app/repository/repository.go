package repository

import (
	"awesomeProject/internal/app/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository struct {
	db             *gorm.DB
	minio          *minio.Client
	minioBucket    string
	minioPublicURL string
}

func New(dsn string, cfg *config.Config) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	minioClient, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}

	return &Repository{
		db:             db,
		minio:          minioClient,
		minioBucket:    cfg.MinioBucket,
		minioPublicURL: cfg.MinioPublicURL,
	}, nil
}
