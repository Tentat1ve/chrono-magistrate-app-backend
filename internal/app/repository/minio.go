package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
)

var reNotLatin = regexp.MustCompile(`[^a-z0-9]`)

// UploadFile кладёт файл в бакет Minio под сгенерированным латинским именем
// и возвращает публичный url объекта
func (r *Repository) UploadFile(ctx context.Context, prefix, originalName, contentType string, file io.Reader, size int64) (string, error) {
	ext := reNotLatin.ReplaceAllString(strings.ToLower(filepath.Ext(originalName)), "")

	randomPart := make([]byte, 4)
	if _, err := rand.Read(randomPart); err != nil {
		return "", err
	}
	objectName := fmt.Sprintf("%s_%d_%s", prefix, time.Now().Unix(), hex.EncodeToString(randomPart))
	if ext != "" {
		objectName += "." + ext
	}

	_, err := r.minio.PutObject(ctx, r.minioBucket, objectName, file, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("не удалось загрузить файл в Minio: %w", err)
	}

	return r.minioPublicURL + "/" + objectName, nil
}

// RemoveFile удаляет объект из Minio по его публичному url (откат при ошибке записи в БД)
func (r *Repository) RemoveFile(ctx context.Context, url string) {
	objectName := strings.TrimPrefix(url, r.minioPublicURL+"/")
	_ = r.minio.RemoveObject(ctx, r.minioBucket, objectName, minio.RemoveObjectOptions{})
}
