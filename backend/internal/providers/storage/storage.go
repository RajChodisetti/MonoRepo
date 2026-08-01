// Package storage provides durable object-storage adapters for restaurant-
// owned and separately licensed media. Google Places content must not be put
// through this package.
package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	platformconfig "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
)

var ErrDisabled = errors.New("object storage is disabled")

type Store interface {
	Configured() bool
	Put(ctx context.Context, key, contentType string, body io.Reader, size int64) error
	Delete(ctx context.Context, key string) error
	PublicURL(key string) string
}

type Disabled struct{}

func (Disabled) Configured() bool { return false }
func (Disabled) Put(context.Context, string, string, io.Reader, int64) error {
	return ErrDisabled
}
func (Disabled) Delete(context.Context, string) error { return ErrDisabled }
func (Disabled) PublicURL(string) string              { return "" }

type s3API interface {
	PutObject(context.Context, *s3.PutObjectInput, ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	DeleteObject(context.Context, *s3.DeleteObjectInput, ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

type S3 struct {
	client        s3API
	bucket        string
	publicBaseURL string
}

func New(ctx context.Context, cfg platformconfig.StorageConfig) (Store, error) {
	provider := strings.ToLower(strings.TrimSpace(cfg.Provider))
	if provider == "" || provider == "disabled" {
		return Disabled{}, nil
	}
	if provider != "s3" {
		return nil, fmt.Errorf("unsupported storage provider %q", provider)
	}
	region := strings.TrimSpace(cfg.Region)
	if region == "" {
		// SigV4 still requires a signing region for S3-compatible endpoints.
		region = "us-east-1"
	}
	loadOptions := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			strings.TrimSpace(cfg.AccessKeyID),
			strings.TrimSpace(cfg.SecretAccessKey),
			"",
		)),
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, loadOptions...)
	if err != nil {
		return nil, fmt.Errorf("load s3 configuration: %w", err)
	}
	client := s3.NewFromConfig(awsCfg, func(options *s3.Options) {
		if endpoint := strings.TrimRight(strings.TrimSpace(cfg.Endpoint), "/"); endpoint != "" {
			options.BaseEndpoint = aws.String(endpoint)
		}
		options.UsePathStyle = cfg.UsePathStyle
	})
	return &S3{
		client:        client,
		bucket:        strings.TrimSpace(cfg.Bucket),
		publicBaseURL: strings.TrimRight(strings.TrimSpace(cfg.PublicBaseURL), "/"),
	}, nil
}

func (store *S3) Configured() bool { return store != nil && store.client != nil }

func (store *S3) Put(
	ctx context.Context,
	key, contentType string,
	body io.Reader,
	size int64,
) error {
	key = cleanObjectKey(key)
	if key == "" {
		return errors.New("object key is required")
	}
	if size <= 0 {
		return errors.New("object size must be positive")
	}
	_, err := store.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(store.bucket),
		Key:           aws.String(key),
		Body:          body,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
		CacheControl:  aws.String("public, max-age=31536000, immutable"),
	})
	if err != nil {
		return fmt.Errorf("put restaurant media object: %w", err)
	}
	return nil
}

func (store *S3) Delete(ctx context.Context, key string) error {
	key = cleanObjectKey(key)
	if key == "" {
		return errors.New("object key is required")
	}
	_, err := store.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(store.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete restaurant media object: %w", err)
	}
	return nil
}

func (store *S3) PublicURL(key string) string {
	key = cleanObjectKey(key)
	if key == "" || store.publicBaseURL == "" {
		return ""
	}
	base, err := url.Parse(store.publicBaseURL + "/")
	if err != nil {
		return ""
	}
	basePath := strings.TrimRight(base.Path, "/")
	escapedBasePath := strings.TrimRight(base.EscapedPath(), "/")
	base.Path = basePath + "/" + key
	base.RawPath = escapedBasePath + "/" + escapeObjectKey(key)
	return base.String()
}

func cleanObjectKey(key string) string {
	raw := strings.TrimSpace(key)
	for _, part := range strings.Split(strings.ReplaceAll(raw, "\\", "/"), "/") {
		if part == ".." {
			return ""
		}
	}
	cleaned := strings.TrimPrefix(path.Clean("/"+raw), "/")
	if cleaned == "." || strings.HasPrefix(cleaned, "../") {
		return ""
	}
	return cleaned
}

func escapeObjectKey(key string) string {
	parts := strings.Split(key, "/")
	for index, part := range parts {
		parts[index] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}

var _ Store = Disabled{}
var _ Store = (*S3)(nil)
