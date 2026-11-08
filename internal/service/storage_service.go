package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"proem-backend/internal/config"
	"proem-backend/internal/domain"
	"strings"

	"github.com/google/uuid"
)

type StorageService interface {
	UploadImage(fileHeader *multipart.FileHeader) (string, error)
}

type localStorageService struct {
	cfg *config.Config
}

const maxFileSize = 5 << 20 // 5MB

func NewLocalStorageService(cfg *config.Config) (StorageService, error) {
	if err := os.MkdirAll(cfg.UploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}
	return &localStorageService{cfg: cfg}, nil
}

func (s *localStorageService) UploadImage(fileHeader *multipart.FileHeader) (string, error) {
	if fileHeader.Size > maxFileSize {
		return "", domain.ErrFileSizeTooLarge
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Валидация MIME-типа через чтение первых 512 байт (magic bytes)
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}
	mimeType := http.DetectContentType(buf[:n])

	allowedMimeTypes := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
		"image/gif":  ".gif",
	}

	ext, allowed := allowedMimeTypes[mimeType]
	if !allowed {
		return "", domain.ErrInvalidFileFormat
	}

	// Сбрасываем курсор чтения файла обратно в начало
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	// Генерируем уникальное имя файла
	newFileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	dstPath := filepath.Join(s.cfg.UploadDir, newFileName)

	dst, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	// Возвращаем публичный путь
	publicURL := fmt.Sprintf("%s/uploads/%s", strings.TrimRight(s.cfg.AppBaseUrl, "/"), newFileName)
	return publicURL, nil
}
