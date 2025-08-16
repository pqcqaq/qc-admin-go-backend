package s3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"go-backend/pkg/configs"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// LoggerInterface 定义日志接口，避免循环依赖
type LoggerInterface interface {
	Info(format string, args ...any)
	Error(format string, args ...any)
	Fatal(format string, args ...any)
}

// S3Client S3客户端结构
type S3Client struct {
	client   *s3.Client
	uploader *manager.Uploader
	config   *configs.S3Config
}

// 单例相关变量
var (
	Client *S3Client
	once   sync.Once
	mu     sync.RWMutex
)

// 全局logger实例
var logger LoggerInterface

// SetLogger 设置logger实例
func SetLogger(l LoggerInterface) {
	logger = l
}

// NewClient 创建新的S3客户端
func NewClient(s3Config *configs.S3Config) (*S3Client, error) {
	ctx := context.Background()

	// 创建AWS配置
	var cfg aws.Config
	var err error

	// 设置凭证
	if s3Config.AccessKeyID != "" && s3Config.SecretAccessKey != "" {
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(s3Config.Region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				s3Config.AccessKeyID,
				s3Config.SecretAccessKey,
				s3Config.SessionToken,
			)),
		)
	} else {
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(s3Config.Region),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// 自定义选项
	var options []func(*s3.Options)

	// 设置端点（用于MinIO等兼容S3的服务）
	if s3Config.Endpoint != "" {
		options = append(options, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(s3Config.Endpoint)
			o.UsePathStyle = s3Config.ForcePathStyle
		})
	}

	// 创建S3客户端
	s3Client := s3.NewFromConfig(cfg, options...)
	uploader := manager.NewUploader(s3Client)

	client := &S3Client{
		client:   s3Client,
		uploader: uploader,
		config:   s3Config,
	}

	// 测试连接
	if err := client.TestConnection(); err != nil {
		if logger != nil {
			logger.Error("S3 connection test failed: %v", err)
		}
		return nil, fmt.Errorf("S3 connection test failed: %w", err)
	}

	if logger != nil {
		logger.Info("S3 client connected successfully to region: %s", s3Config.Region)
	}

	return client, nil
}

// GetClient 获取S3客户端单例
func GetClient() *S3Client {
	mu.RLock()
	defer mu.RUnlock()
	return Client
}

// InitClient 初始化S3客户端单例
func InitClient(config *configs.S3Config) error {
	var err error
	once.Do(func() {
		mu.Lock()
		defer mu.Unlock()
		Client, err = NewClient(config)
	})
	return err
}

// TestConnection 测试S3连接
func (c *S3Client) TestConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 尝试列出存储桶来测试连接
	_, err := c.client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return fmt.Errorf("failed to connect to S3: %w", err)
	}

	return nil
}

// UploadFile 上传文件到S3
func (c *S3Client) UploadFile(bucket, key string, body io.Reader, contentType string) (*manager.UploadOutput, error) {
	if bucket == "" {
		bucket = c.config.Bucket
	}

	ctx := context.Background()
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   body,
	}

	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}

	return c.uploader.Upload(ctx, input)
}

// DownloadFile 从S3下载文件
func (c *S3Client) DownloadFile(bucket, key string, writer io.WriterAt) (int64, error) {
	if bucket == "" {
		bucket = c.config.Bucket
	}

	ctx := context.Background()
	downloader := manager.NewDownloader(c.client)

	return downloader.Download(ctx, writer, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
}

// DeleteFile 从S3删除文件
func (c *S3Client) DeleteFile(bucket, key string) error {
	if bucket == "" {
		bucket = c.config.Bucket
	}

	ctx := context.Background()
	_, err := c.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	return err
}

// GetFileURL 获取文件的预签名URL
func (c *S3Client) GetFileURL(bucket, key string, expiration time.Duration) (string, error) {
	if bucket == "" {
		bucket = c.config.Bucket
	}

	ctx := context.Background()
	presignClient := s3.NewPresignClient(c.client)

	req, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiration
	})

	if err != nil {
		return "", err
	}

	return req.URL, nil
}

// GetPresignedPutURL 获取文件上传的预签名URL
func (c *S3Client) GetPresignedPutURL(bucket, key string, expiration time.Duration, contentType string) (string, error) {
	if bucket == "" {
		bucket = c.config.Bucket
	}

	ctx := context.Background()
	presignClient := s3.NewPresignClient(c.client)

	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}

	req, err := presignClient.PresignPutObject(ctx, input, func(opts *s3.PresignOptions) {
		opts.Expires = expiration
	})

	if err != nil {
		return "", err
	}

	return req.URL, nil
}

// FileExists 检查文件是否存在
func (c *S3Client) FileExists(bucket, key string) (bool, error) {
	if bucket == "" {
		bucket = c.config.Bucket
	}

	ctx := context.Background()
	_, err := c.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		if IsNotFoundError(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// GetFileInfo 获取文件信息
func (c *S3Client) GetFileInfo(bucket, key string) (*s3.HeadObjectOutput, error) {
	if bucket == "" {
		bucket = c.config.Bucket
	}

	ctx := context.Background()
	return c.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
}

// ListFiles 列出文件
func (c *S3Client) ListFiles(bucket, prefix string, maxKeys int32) ([]types.Object, error) {
	if bucket == "" {
		bucket = c.config.Bucket
	}

	ctx := context.Background()
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	}

	if prefix != "" {
		input.Prefix = aws.String(prefix)
	}

	if maxKeys > 0 {
		input.MaxKeys = aws.Int32(maxKeys)
	}

	result, err := c.client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, err
	}

	return result.Contents, nil
}

// CopyFile 复制文件
func (c *S3Client) CopyFile(sourceBucket, sourceKey, destBucket, destKey string) error {
	if sourceBucket == "" {
		sourceBucket = c.config.Bucket
	}
	if destBucket == "" {
		destBucket = c.config.Bucket
	}

	source := fmt.Sprintf("%s/%s", sourceBucket, sourceKey)

	ctx := context.Background()
	_, err := c.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(destBucket),
		Key:        aws.String(destKey),
		CopySource: aws.String(source),
	})

	return err
}

// CreateBucket 创建存储桶
func (c *S3Client) CreateBucket(bucket string) error {
	ctx := context.Background()
	_, err := c.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	return err
}

// DeleteBucket 删除存储桶
func (c *S3Client) DeleteBucket(bucket string) error {
	ctx := context.Background()
	_, err := c.client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(bucket),
	})
	return err
}

// BucketExists 检查存储桶是否存在
func (c *S3Client) BucketExists(bucket string) (bool, error) {
	ctx := context.Background()
	_, err := c.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})

	if err != nil {
		if IsNotFoundError(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// IsNotFoundError 检查是否为404错误
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	// 检查是否为NoSuchKey或NoSuchBucket错误
	var noSuchKey *types.NoSuchKey
	var noSuchBucket *types.NoSuchBucket
	var notFound *types.NotFound

	if errors.As(err, &noSuchKey) ||
		errors.As(err, &noSuchBucket) ||
		errors.As(err, &notFound) {
		return true
	}

	return false
}
