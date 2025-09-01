package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOStorage gerencia operações com MinIO
type MinIOStorage struct {
	client     *minio.Client
	bucketName string
	baseURL    string
}

// NewMinIOStorage cria uma nova instância do MinIO storage
func NewMinIOStorage(endpoint, accessKey, secretKey, bucketName, baseURL string, useSSL bool) (*MinIOStorage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cliente MinIO: %w", err)
	}

	storage := &MinIOStorage{
		client:     client,
		bucketName: bucketName,
		baseURL:    baseURL,
	}

	// Verificar se o bucket existe, se não criar
	if err := storage.ensureBucket(context.Background()); err != nil {
		return nil, fmt.Errorf("erro ao verificar/criar bucket: %w", err)
	}

	return storage, nil
}

// ensureBucket garante que o bucket existe
func (s *MinIOStorage) ensureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucketName)
	if err != nil {
		return err
	}

	if !exists {
		return s.client.MakeBucket(ctx, s.bucketName, minio.MakeBucketOptions{})
	}

	return nil
}

// GetPresignedUploadURL gera uma URL pré-assinada para upload
func (s *MinIOStorage) GetPresignedUploadURL(ctx context.Context, objectName, contentType string, expiry int64) (string, error) {
	presignedURL, err := s.client.PresignedPutObject(ctx, s.bucketName, objectName, time.Duration(expiry)*time.Second)
	if err != nil {
		return "", fmt.Errorf("erro ao gerar URL pré-assinada para upload: %w", err)
	}

	return presignedURL.String(), nil
}

// GetPresignedDownloadURL gera uma URL pré-assinada para download
func (s *MinIOStorage) GetPresignedDownloadURL(ctx context.Context, objectName string, expiry int64) (string, error) {
	presignedURL, err := s.client.PresignedGetObject(ctx, s.bucketName, objectName, time.Duration(expiry)*time.Second, nil)
	if err != nil {
		return "", fmt.Errorf("erro ao gerar URL pré-assinada para download: %w", err)
	}

	return presignedURL.String(), nil
}

// GetPublicURL retorna a URL pública de um objeto
func (s *MinIOStorage) GetPublicURL(objectName string) string {
	return fmt.Sprintf("%s/%s/%s", s.baseURL, s.bucketName, objectName)
}

// FileExists verifica se um arquivo existe no storage
func (s *MinIOStorage) FileExists(ctx context.Context, objectName string) (bool, error) {
	_, err := s.client.StatObject(ctx, s.bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		// Se o erro for de objeto não encontrado, retorna false
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("erro ao verificar existência do arquivo: %w", err)
	}
	return true, nil
}

// DeleteObject remove um objeto do storage
func (s *MinIOStorage) DeleteObject(ctx context.Context, objectName string) error {
	return s.client.RemoveObject(ctx, s.bucketName, objectName, minio.RemoveObjectOptions{})
}

// ListObjects lista objetos com determinado prefixo
func (s *MinIOStorage) ListObjects(ctx context.Context, prefix string) (<-chan minio.ObjectInfo, error) {
	return s.client.ListObjects(ctx, s.bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	}), nil
}

// GetObjectInfo obtém informações sobre um objeto
func (s *MinIOStorage) GetObjectInfo(ctx context.Context, objectName string) (*minio.ObjectInfo, error) {
	info, err := s.client.StatObject(ctx, s.bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("erro ao obter informações do objeto: %w", err)
	}
	return &info, nil
}

// CopyObject copia um objeto dentro do storage
func (s *MinIOStorage) CopyObject(ctx context.Context, sourceObjectName, destObjectName string) error {
	src := minio.CopySrcOptions{
		Bucket: s.bucketName,
		Object: sourceObjectName,
	}

	dest := minio.CopyDestOptions{
		Bucket: s.bucketName,
		Object: destObjectName,
	}

	_, err := s.client.CopyObject(ctx, dest, src)
	if err != nil {
		return fmt.Errorf("erro ao copiar objeto: %w", err)
	}

	return nil
}

// GenerateZipDownloadURL cria um arquivo ZIP temporário e retorna URL de download
func (s *MinIOStorage) GenerateZipDownloadURL(ctx context.Context, objects []string, zipFileName string) (string, error) {
	// TODO: Implementar geração de ZIP
	// Esta funcionalidade seria mais complexa e requereria:
	// 1. Criar um ZIP temporário com os arquivos selecionados
	// 2. Fazer upload do ZIP para o MinIO
	// 3. Retornar URL pré-assinada para download do ZIP
	// 4. Configurar limpeza automática do ZIP após expiração

	return "", fmt.Errorf("geração de ZIP ainda não implementada")
}