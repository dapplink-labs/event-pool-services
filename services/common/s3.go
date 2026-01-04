package common

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"

	appconfig "github.com/multimarket-labs/event-pod-services/config"
)

type S3Service struct {
	client *s3.Client
	config *appconfig.S3Config
}

func NewS3Service(s3Config *appconfig.S3Config) (*S3Service, error) {
	if s3Config.AccessKey == "" || s3Config.SecretKey == "" {
		return nil, fmt.Errorf("aws access key id and secret access key are required")
	}
	if s3Config.Bucket == "" {
		return nil, fmt.Errorf("aws s3 bucket is required")
	}
	if s3Config.Region == "" {
		return nil, fmt.Errorf("aws s3 region is required")
	}

	creds := credentials.NewStaticCredentialsProvider(
		s3Config.AccessKey,
		s3Config.SecretKey,
		"",
	)

	var cfg aws.Config
	var err error

	if s3Config.Endpoint != "" {
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:               s3Config.Endpoint,
				SigningRegion:     s3Config.Region,
				HostnameImmutable: true,
			}, nil
		})

		cfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(s3Config.Region),
			config.WithCredentialsProvider(creds),
			config.WithEndpointResolverWithOptions(customResolver),
		)
	} else {
		cfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(s3Config.Region),
			config.WithCredentialsProvider(creds),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = s3Config.UsePathStyle
	})

	return &S3Service{
		client: client,
		config: s3Config,
	}, nil
}

func (s *S3Service) UploadFile(ctx context.Context, fileData []byte, fileName string) (string, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(fileName),
		Body:   bytes.NewReader(fileData),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	fileURL := s.GetPublicURL(fileName)
	return fileURL, nil
}

func (s *S3Service) UploadStream(ctx context.Context, reader io.Reader, fileName string, fileSize int64) (string, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.config.Bucket),
		Key:           aws.String(fileName),
		Body:          reader,
		ContentLength: aws.Int64(fileSize),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload stream: %w", err)
	}

	fileURL := s.GetPublicURL(fileName)
	return fileURL, nil
}

func (s *S3Service) UploadFileWithContentType(ctx context.Context, fileData []byte, fileName string, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.config.Bucket),
		Key:         aws.String(fileName),
		Body:        bytes.NewReader(fileData),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	fileURL := s.GetPublicURL(fileName)
	return fileURL, nil
}

func (s *S3Service) DeleteFile(ctx context.Context, fileName string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(fileName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (s *S3Service) GetFileInfo(ctx context.Context, fileName string) (*s3.HeadObjectOutput, error) {
	result, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(fileName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	return result, nil
}

func (s *S3Service) GetPresignedURL(ctx context.Context, fileName string, expireSeconds int64) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(fileName),
	}, s3.WithPresignExpires(time.Duration(expireSeconds)*time.Second))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned url: %w", err)
	}

	return request.URL, nil
}

func (s *S3Service) GetPresignedUploadURL(ctx context.Context, fileName string, expireSeconds int64) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	request, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(fileName),
	}, s3.WithPresignExpires(time.Duration(expireSeconds)*time.Second))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned upload url: %w", err)
	}

	return request.URL, nil
}

func (s *S3Service) GetPublicURL(fileName string) string {
	if s.config.CDNDomain != "" {
		return fmt.Sprintf("%s/%s", s.config.CDNDomain, fileName)
	}
	if s.config.Endpoint != "" {
		return fmt.Sprintf("%s/%s/%s", s.config.Endpoint, s.config.Bucket, fileName)
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.config.Bucket, s.config.Region, fileName)
}

func (s *S3Service) ListFiles(ctx context.Context, prefix string, maxKeys int32) ([]string, error) {
	if maxKeys <= 0 {
		maxKeys = 100
	}

	var files []string
	var continuationToken *string

	for {
		input := &s3.ListObjectsV2Input{
			Bucket:            aws.String(s.config.Bucket),
			Prefix:            aws.String(prefix),
			MaxKeys:           aws.Int32(maxKeys),
			ContinuationToken: continuationToken,
		}

		result, err := s.client.ListObjectsV2(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to list files: %w", err)
		}

		for _, obj := range result.Contents {
			files = append(files, *obj.Key)
		}

		if !*result.IsTruncated {
			break
		}
		continuationToken = result.NextContinuationToken
	}

	return files, nil
}

func (s *S3Service) CopyFile(ctx context.Context, srcFileName, destFileName string) error {
	copySource := fmt.Sprintf("%s/%s", s.config.Bucket, srcFileName)

	_, err := s.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(s.config.Bucket),
		CopySource: aws.String(copySource),
		Key:        aws.String(destFileName),
	})
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}
	return nil
}

func (s *S3Service) MoveFile(ctx context.Context, srcFileName, destFileName string) error {
	err := s.CopyFile(ctx, srcFileName, destFileName)
	if err != nil {
		return err
	}

	err = s.DeleteFile(ctx, srcFileName)
	if err != nil {
		return fmt.Errorf("failed to delete source file after copy: %w", err)
	}
	return nil
}

func (s *S3Service) BatchDelete(ctx context.Context, fileNames []string) error {
	var objects []s3types.ObjectIdentifier
	for _, fileName := range fileNames {
		objects = append(objects, s3types.ObjectIdentifier{
			Key: aws.String(fileName),
		})
	}

	_, err := s.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(s.config.Bucket),
		Delete: &s3types.Delete{
			Objects: objects,
			Quiet:   aws.Bool(true),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to batch delete: %w", err)
	}

	return nil
}

func (s *S3Service) DownloadFile(ctx context.Context, fileName string) ([]byte, error) {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(fileName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	return data, nil
}

func (s *S3Service) FileExists(ctx context.Context, fileName string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(fileName),
	})
	if err != nil {
		return false, nil
	}
	return true, nil
}
