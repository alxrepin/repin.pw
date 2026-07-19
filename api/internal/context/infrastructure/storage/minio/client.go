package minio

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"repin/internal/context/domain"
)

type Client struct {
	client *minio.Client
	bucket string
}

// Object is a stored file opened for reading. Body seeks, so callers can serve
// range requests straight from it.
type Object struct {
	Body        io.ReadSeekCloser
	ContentType string
	ModTime     time.Time
	Size        int64
}

func NewClient(endpoint, accessKey, secretKey, bucket string) (*Client, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, fmt.Errorf("minio client: %w", err)
	}

	return &Client{client: client, bucket: bucket}, nil
}

// Upload returns the object key rather than a URL: the bucket and endpoint are
// deployment details, and baking them into stored rows outlives any redeploy.
func (c *Client) Upload(ctx context.Context, objectName string, data []byte, contentType string) (string, error) {
	_, err := c.client.PutObject(ctx, c.bucket, objectName, bytes.NewReader(data), int64(len(data)),
		minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", fmt.Errorf("put object: %w", err)
	}

	return objectName, nil
}

func (c *Client) Get(ctx context.Context, objectName string) (*Object, error) {
	obj, err := c.client.GetObject(ctx, c.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get object: %w", err)
	}

	// GetObject is lazy — the request only reaches the server on Stat, so this
	// is where a missing key surfaces.
	info, err := obj.Stat()
	if err != nil {
		_ = obj.Close()

		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return nil, domain.ErrMediaNotFound
		}

		return nil, fmt.Errorf("stat object: %w", err)
	}

	return &Object{
		Body:        obj,
		ContentType: info.ContentType,
		ModTime:     info.LastModified,
		Size:        info.Size,
	}, nil
}
