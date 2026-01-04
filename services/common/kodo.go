package common

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"

	"github.com/multimarket-labs/event-pod-services/config"
)

type KodoService struct {
	config    *config.KodoConfig
	mac       *qbox.Mac
	bucketMgr *storage.BucketManager
	putPolicy storage.PutPolicy
	uploader  *storage.FormUploader
	cfg       storage.Config
}

func NewKodoService(kodoConfig *config.KodoConfig) (*KodoService, error) {
	if kodoConfig.AccessKey == "" || kodoConfig.SecretKey == "" {
		return nil, fmt.Errorf("qiniu access key and secret key are required")
	}
	if kodoConfig.Bucket == "" {
		return nil, fmt.Errorf("qiniu bucket is required")
	}

	mac := qbox.NewMac(kodoConfig.AccessKey, kodoConfig.SecretKey)

	cfg := storage.Config{
		UseHTTPS:      kodoConfig.UseHTTPS,
		UseCdnDomains: kodoConfig.UseCdnDomains,
	}

	switch kodoConfig.Zone {
	case "zone0":
		cfg.Zone = &storage.ZoneHuadong
	case "zone1":
		cfg.Zone = &storage.ZoneHuabei
	case "zone2":
		cfg.Zone = &storage.ZoneHuanan
	case "zone_na0":
		cfg.Zone = &storage.ZoneBeimei
	case "zone_as0":
		cfg.Zone = &storage.ZoneXinjiapo
	default:
		cfg.Zone = &storage.ZoneHuadong
	}

	bucketMgr := storage.NewBucketManager(mac, &cfg)

	putPolicy := storage.PutPolicy{
		Scope: kodoConfig.Bucket,
	}

	uploader := storage.NewFormUploader(&cfg)

	return &KodoService{
		config:    kodoConfig,
		mac:       mac,
		bucketMgr: bucketMgr,
		putPolicy: putPolicy,
		uploader:  uploader,
		cfg:       cfg,
	}, nil
}

func (s *KodoService) UploadFile(ctx context.Context, fileData []byte, fileName string) (string, error) {
	upToken := s.putPolicy.UploadToken(s.mac)

	ret := storage.PutRet{}
	putExtra := storage.PutExtra{}

	dataLen := int64(len(fileData))
	err := s.uploader.Put(ctx, &ret, upToken, fileName, bytes.NewReader(fileData), dataLen, &putExtra)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	fileURL := fmt.Sprintf("%s/%s", s.config.Domain, ret.Key)
	return fileURL, nil
}

func (s *KodoService) UploadStream(ctx context.Context, reader io.Reader, fileName string, fileSize int64) (string, error) {
	upToken := s.putPolicy.UploadToken(s.mac)

	ret := storage.PutRet{}
	putExtra := storage.PutExtra{}

	err := s.uploader.Put(ctx, &ret, upToken, fileName, reader, fileSize, &putExtra)
	if err != nil {
		return "", fmt.Errorf("failed to upload stream: %w", err)
	}

	fileURL := fmt.Sprintf("%s/%s", s.config.Domain, ret.Key)
	return fileURL, nil
}

func (s *KodoService) DeleteFile(fileName string) error {
	err := s.bucketMgr.Delete(s.config.Bucket, fileName)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (s *KodoService) GetFileInfo(fileName string) (*storage.FileInfo, error) {
	fileInfo, err := s.bucketMgr.Stat(s.config.Bucket, fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	return &fileInfo, nil
}

func (s *KodoService) GetPrivateURL(fileName string, expireSeconds int64) string {
	deadline := time.Now().Add(time.Second * time.Duration(expireSeconds)).Unix()
	privateAccessURL := storage.MakePrivateURL(s.mac, s.config.Domain, fileName, deadline)
	return privateAccessURL
}

func (s *KodoService) GetPublicURL(fileName string) string {
	publicAccessURL := storage.MakePublicURL(s.config.Domain, fileName)
	return publicAccessURL
}

func (s *KodoService) ListFiles(prefix string, limit int) ([]storage.ListItem, error) {
	if limit <= 0 {
		limit = 100
	}

	var items []storage.ListItem
	marker := ""

	for {
		entries, _, nextMarker, hasNext, err := s.bucketMgr.ListFiles(s.config.Bucket, prefix, "", marker, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to list files: %w", err)
		}

		items = append(items, entries...)

		if !hasNext {
			break
		}
		marker = nextMarker
	}

	return items, nil
}

func (s *KodoService) CopyFile(srcFileName, destFileName string) error {
	err := s.bucketMgr.Copy(s.config.Bucket, srcFileName, s.config.Bucket, destFileName, true)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}
	return nil
}

func (s *KodoService) MoveFile(srcFileName, destFileName string) error {
	err := s.bucketMgr.Move(s.config.Bucket, srcFileName, s.config.Bucket, destFileName, true)
	if err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}
	return nil
}

func (s *KodoService) BatchDelete(fileNames []string) error {
	deleteOps := make([]string, 0, len(fileNames))
	for _, fileName := range fileNames {
		deleteOps = append(deleteOps, storage.URIDelete(s.config.Bucket, fileName))
	}

	rets, err := s.bucketMgr.Batch(deleteOps)
	if err != nil {
		return fmt.Errorf("failed to batch delete: %w", err)
	}

	for i, ret := range rets {
		if ret.Code != 200 {
			return fmt.Errorf("failed to delete file %s: code=%d", fileNames[i], ret.Code)
		}
	}

	return nil
}
