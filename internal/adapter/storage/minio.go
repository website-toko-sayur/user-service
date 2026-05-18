package storage

import (
	"bytes"
	"context"
	"fmt"
	"time"
	"user-service/config"

	"github.com/minio/minio-go/v7"
)

type MinioStorageStruct struct {
	cfg        *config.Config
	client     *minio.Client
	bucketName string
}

type MinioStorageInterface interface {
	UploadFile(ctx context.Context, path string, file []byte) (string, error)
}

func NewMinioStorage(cfg *config.Config, client *minio.Client, bucket string) MinioStorageInterface {
	return &MinioStorageStruct{
		cfg:        cfg,
		client:     client,
		bucketName: bucket,
	}
}

func (m *MinioStorageStruct) UploadFile(ctx context.Context, path string, file []byte) (string, error) {
	const timeout = 60 * time.Second

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	_, err := m.client.PutObject(
		ctx,
		m.bucketName,
		path,
		bytes.NewReader(file),
		int64(len(file)),
		minio.PutObjectOptions{
			ContentType: "image/jpeg",
		},
	)
	if err != nil {
		return "", err
	}

	// waktu pakai docker, ganti url host
	publicURL := m.cfg.Storage.PublicURL
	url := fmt.Sprintf("%s/%s/%s", publicURL, m.bucketName, path)

	return url, nil
}
