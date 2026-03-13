package storage

import (
	"bytes"
	"context"
	"io"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const opTimeout = 10 * time.Second

type Client struct {
	s3client      *s3.Client
	presignClient *s3.PresignClient
	bucket        string
	presignExpiry time.Duration
}

// NewClient creates an S3 storage client.
// Credentials are resolved automatically: IAM role on EC2, or AWS_ACCESS_KEY_ID/AWS_SECRET_ACCESS_KEY env vars for local dev.
// endpoint is optional — set to MinIO URL (e.g. "http://localhost:9000") for local development.
func NewClient(ctx context.Context, region, bucket, endpoint string, presignExpiry time.Duration) (*Client, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithRequestChecksumCalculation(aws.RequestChecksumCalculationWhenRequired),
	)
	if err != nil {
		return nil, err
	}

	opts := []func(*s3.Options){}
	if endpoint != "" {
		opts = append(opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true // required for MinIO
		})
	}

	client := s3.NewFromConfig(cfg, opts...)

	return &Client{
		s3client:      client,
		presignClient: s3.NewPresignClient(client),
		bucket:        bucket,
		presignExpiry: presignExpiry,
	}, nil
}

func (c *Client) PutObject(ctx context.Context, objectKey, contentType string, r io.Reader, size int64) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	// Avoid double-buffering if the caller already has a seekable reader (e.g. bytes.NewReader).
	rs, ok := r.(io.ReadSeeker)
	if !ok {
		buf, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		size = int64(len(buf))
		rs = bytes.NewReader(buf)
	}
	_, err := c.s3client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(c.bucket),
		Key:           aws.String(objectKey),
		Body:          rs,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
	})
	return err
}

// GetObject downloads an object from S3 and returns its body.
// The caller is responsible for closing the returned ReadCloser.
// No internal timeout is applied — use a context deadline if needed.
func (c *Client) GetObject(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	out, err := c.s3client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return nil, err
	}
	return out.Body, nil
}

func (c *Client) DeleteObject(ctx context.Context, objectKey string) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	_, err := c.s3client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(objectKey),
	})
	return err
}

func (c *Client) PresignedGetURL(ctx context.Context, objectKey string) (*url.URL, error) {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	req, err := c.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(objectKey),
	}, func(o *s3.PresignOptions) {
		o.Expires = c.presignExpiry
	})
	if err != nil {
		return nil, err
	}
	return url.Parse(req.URL)
}
