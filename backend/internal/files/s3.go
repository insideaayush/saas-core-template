package files

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Provider struct {
	bucket  string
	client  *s3.Client
	presign *s3.PresignClient
}

type S3Config struct {
	Bucket          string
	Region          string
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	ForcePathStyle  bool
}

func NewS3Provider(ctx context.Context, cfg S3Config) (*S3Provider, error) {
	bucket := strings.TrimSpace(cfg.Bucket)
	if bucket == "" {
		return nil, fmt.Errorf("missing S3_BUCKET")
	}

	loadOpts := []func(*config.LoadOptions) error{
		config.WithRegion(strings.TrimSpace(defaultString(cfg.Region, "auto"))),
	}

	if strings.TrimSpace(cfg.AccessKeyID) != "" && strings.TrimSpace(cfg.SecretAccessKey) != "" {
		loadOpts = append(loadOpts, config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			strings.TrimSpace(cfg.AccessKeyID),
			strings.TrimSpace(cfg.SecretAccessKey),
			"",
		)))
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, loadOpts...)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	clientOpts := []func(*s3.Options){}
	if endpoint := strings.TrimSpace(cfg.Endpoint); endpoint != "" {
		clientOpts = append(clientOpts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
		})
	}
	if cfg.ForcePathStyle {
		clientOpts = append(clientOpts, func(o *s3.Options) {
			o.UsePathStyle = true
		})
	}

	s3Client := s3.NewFromConfig(awsCfg, clientOpts...)
	p := s3.NewPresignClient(s3Client)

	return &S3Provider{bucket: bucket, client: s3Client, presign: p}, nil
}

func (p *S3Provider) PresignPut(ctx context.Context, key string, contentType string, ttl time.Duration) (string, map[string]string, error) {
	if strings.TrimSpace(key) == "" {
		return "", nil, fmt.Errorf("missing key")
	}
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}

	out, err := p.presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(p.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", nil, fmt.Errorf("presign put: %w", err)
	}

	return out.URL, map[string]string{"Content-Type": contentType}, nil
}

func (p *S3Provider) PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error) {
	if strings.TrimSpace(key) == "" {
		return "", fmt.Errorf("missing key")
	}
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}

	out, err := p.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", fmt.Errorf("presign get: %w", err)
	}

	return out.URL, nil
}
