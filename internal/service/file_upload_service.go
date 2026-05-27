package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// FileUploadService сервис для загрузки файлов
type FileUploadService struct {
	uploadDir string
	baseURL   string
}

// NewFileUploadService создает новый FileUploadService
func NewFileUploadService(uploadDir string, baseURL string) *FileUploadService {
	// Создаем директорию для загрузок если не существует
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		panic(fmt.Sprintf("failed to create upload directory: %v", err))
	}

	return &FileUploadService{
		uploadDir: uploadDir,
		baseURL:   baseURL,
	}
}

// UploadResult результат загрузки файла
type UploadResult struct {
	URL      string `json:"url"`
	FilePath string `json:"file_path"`
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
	MimeType string `json:"mime_type"`
}

// UploadAvatar загружает аватар пользователя
func (s *FileUploadService) UploadAvatar(ctx context.Context, file *multipart.FileHeader, userID int64) (*UploadResult, error) {
	return s.uploadFile(ctx, file, "avatars", userID)
}

// UploadBanner загружает фон профиля пользователя
func (s *FileUploadService) UploadBanner(ctx context.Context, file *multipart.FileHeader, userID int64) (*UploadResult, error) {
	return s.uploadFile(ctx, file, "banners", userID)
}

// uploadFile загружает файл в указанную поддиректорию
func (s *FileUploadService) uploadFile(ctx context.Context, file *multipart.FileHeader, subdir string, userID int64) (*UploadResult, error) {
	// Открываем файл
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// Проверяем MIME тип
	mimeType := file.Header.Get("Content-Type")
	if !isValidImageType(mimeType) {
		return nil, fmt.Errorf("invalid file type: %s (allowed: image/jpeg, image/png, image/webp, image/gif)", mimeType)
	}

	// Проверяем размер (макс 5MB)
	if file.Size > 5*1024*1024 {
		return nil, fmt.Errorf("file too large: %d bytes (max: 5MB)", file.Size)
	}

	// Создаем уникальное имя файла
	ext := getFileExtension(file.Filename, mimeType)
	fileName := fmt.Sprintf("%d_%s%s", userID, uuid.New().String(), ext)

	// Создаем директорию если не существует
	dir := filepath.Join(s.uploadDir, subdir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Путь для сохранения
	filePath := filepath.Join(dir, fileName)

	// Создаем файл для записи
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// Копируем содержимое
	fileSize, err := io.Copy(dst, src)
	if err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Формируем URL
	url := strings.TrimRight(s.baseURL, "/") + "/uploads/" + subdir + "/" + fileName

	return &UploadResult{
		URL:      url,
		FilePath: filePath,
		FileName: fileName,
		FileSize: fileSize,
		MimeType: mimeType,
	}, nil
}

// isValidImageType проверяет допустимый MIME тип
func isValidImageType(mimeType string) bool {
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
		"image/gif":  true,
	}
	return allowedTypes[mimeType]
}

// getFileExtension получает расширение файла
func getFileExtension(filename, mimeType string) string {
	// Пробуем получить из имени файла
	if ext := filepath.Ext(filename); ext != "" {
		return ext
	}

	// Определяем по MIME типу
	switch mimeType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ".bin"
	}
}

// DeleteFile удаляет файл
func (s *FileUploadService) DeleteFile(ctx context.Context, filePath string) error {
	// Извлекаем относительный путь из URL
	relativePath := strings.TrimPrefix(filePath, s.baseURL)
	relativePath = strings.TrimPrefix(relativePath, "/uploads/")

	// Полный путь к файлу
	fullPath := filepath.Join(s.uploadDir, relativePath)

	// Проверяем существование
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil // Файл уже не существует
	}

	return os.Remove(fullPath)
}

// CleanupOldFiles удаляет файлы старше указанного времени
func (s *FileUploadService) CleanupOldFiles(ctx context.Context, olderThan time.Duration) error {
	now := time.Now()
	cutoff := now.Add(-olderThan)

	return filepath.Walk(s.uploadDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if info.ModTime().Before(cutoff) {
			return os.Remove(path)
		}

		return nil
	})
}
