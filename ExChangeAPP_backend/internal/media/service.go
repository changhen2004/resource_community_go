package media

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	CoverMaxSizeBytes        int64 = 2 << 20
	ContentImageMaxSizeBytes int64 = 5 << 20
	ContentImageMaxCount           = 6
	coverSubDir                    = "covers"
	contentSubDir                  = "content"
)

var (
	ErrFileRequired    = errors.New("file is required")
	ErrTooManyFiles    = errors.New("too many files")
	ErrFileTooLarge    = errors.New("file exceeds size limit")
	ErrUnsupportedType = errors.New("unsupported image type")
)

var allowedContentTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

type Service struct {
	uploadRootDir string
}

func NewService(uploadRootDir string) *Service {
	if uploadRootDir == "" {
		uploadRootDir = "uploads"
	}
	return &Service{uploadRootDir: uploadRootDir}
}

func (s *Service) SaveCover(fileHeader *multipart.FileHeader) (string, error) {
	if fileHeader == nil {
		return "", ErrFileRequired
	}
	return s.saveImage(fileHeader, coverSubDir, CoverMaxSizeBytes)
}

func (s *Service) SaveContentImages(fileHeaders []*multipart.FileHeader) ([]string, error) {
	if len(fileHeaders) == 0 {
		return nil, ErrFileRequired
	}
	if len(fileHeaders) > ContentImageMaxCount {
		return nil, ErrTooManyFiles
	}

	urls := make([]string, 0, len(fileHeaders))
	for _, fileHeader := range fileHeaders {
		url, err := s.saveImage(fileHeader, contentSubDir, ContentImageMaxSizeBytes)
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}
	return urls, nil
}

func (s *Service) saveImage(fileHeader *multipart.FileHeader, subDir string, maxSize int64) (string, error) {
	if fileHeader.Size > maxSize {
		return "", ErrFileTooLarge
	}

	src, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	sniffBuffer := make([]byte, 512)
	n, err := io.ReadFull(src, sniffBuffer)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return "", err
	}

	contentType := http.DetectContentType(sniffBuffer[:n])
	ext, ok := allowedContentTypes[contentType]
	if !ok {
		return "", ErrUnsupportedType
	}

	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	dirPath := filepath.Join(s.uploadRootDir, subDir)
	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		return "", err
	}

	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), sanitizeFilename(fileHeader.Filename), ext)
	dstPath := filepath.Join(dirPath, filename)

	dst, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	return "/" + filepath.ToSlash(dstPath), nil
}

func sanitizeFilename(name string) string {
	base := strings.TrimSuffix(filepath.Base(name), filepath.Ext(name))
	base = strings.ToLower(strings.TrimSpace(base))
	base = strings.ReplaceAll(base, " ", "_")
	if base == "" {
		return "image"
	}
	return base
}
