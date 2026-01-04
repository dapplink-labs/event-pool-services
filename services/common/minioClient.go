package common

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/multimarket-labs/event-pod-services/config"
)

type StorageService struct {
	Client *minio.Client
	Config config.MinioConfig
}

func NewStorageService(cfg config.MinioConfig) *StorageService {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		log.Fatalf("init MinIO fail: %v", err)
	}
	return &StorageService{Client: client, Config: cfg}
}

func (s *StorageService) UploadFile(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader, folder string) (string, error) {
	exists, err := s.Client.BucketExists(ctx, s.Config.BucketName)
	if err != nil {
		return "", fmt.Errorf("check bucket fail: %w", err)
	}
	if !exists {
		err = s.Client.MakeBucket(ctx, s.Config.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return "", fmt.Errorf("create bucket fail: %w", err)
		}
	}

	ext := filepath.Ext(fileHeader.Filename)
	if ext == "" {
		ext = ".dat"
	}
	objectName := fmt.Sprintf("%s/%s/%s%s",
		strings.Trim(folder, "/"),
		time.Now().Format("20060102"),
		uuid.New().String(),
		ext,
	)

	_, err = s.Client.PutObject(ctx,
		s.Config.BucketName,
		objectName,
		file,
		fileHeader.Size,
		minio.PutObjectOptions{ContentType: fileHeader.Header.Get("Content-Type")},
	)
	if err != nil {
		return "", fmt.Errorf("upload file fail: %w", err)
	}

	storagePath := fmt.Sprintf("/%s/%s", s.Config.BucketName, objectName)
	return storagePath, nil
}

func (s *StorageService) GenerateAccessURL(storagePath string) string {
	return fmt.Sprintf("%s%s", strings.TrimRight(s.Config.BaseURL, "/"), storagePath)
}

func (s *StorageService) GenerateAccessURLPrivate(ctx context.Context, objectName string) (string, error) {
	url, err := s.Client.PresignedGetObject(
		ctx,
		s.Config.BucketName,
		objectName,
		time.Hour*24,
		nil,
	)
	return url.String(), err
}
