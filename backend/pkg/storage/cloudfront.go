package storage

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/url"
	"time"

	cfsign "github.com/aws/aws-sdk-go-v2/feature/cloudfront/sign"
)

// CloudFrontSigner generates CloudFront signed URLs for S3 objects served via a
// CloudFront distribution. It satisfies the same PresignedGetURL signature as
// storage.Client so it can be used as a drop-in via the urlcache.Signer interface.
type CloudFrontSigner struct {
	signer    *cfsign.URLSigner
	domain    string
	expiry    time.Duration
}

// NewCloudFrontSigner parses the PEM-encoded RSA private key and returns a signer.
func NewCloudFrontSigner(domain, keyPairID, privateKeyPEM string, expiry time.Duration) (*CloudFrontSigner, error) {
	key, err := parseRSAPrivateKey(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("cloudfront: parse private key: %w", err)
	}
	return &CloudFrontSigner{
		signer: cfsign.NewURLSigner(keyPairID, key),
		domain: domain,
		expiry: expiry,
	}, nil
}

// PresignedGetURL returns a CloudFront signed URL for the given object key.
func (c *CloudFrontSigner) PresignedGetURL(_ context.Context, objectKey string) (*url.URL, error) {
	raw := fmt.Sprintf("https://%s/%s", c.domain, objectKey)
	signed, err := c.signer.Sign(raw, time.Now().Add(c.expiry))
	if err != nil {
		return nil, fmt.Errorf("cloudfront: sign url: %w", err)
	}
	return url.Parse(signed)
}

func parseRSAPrivateKey(pemStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("no PEM block found")
	}
	switch block.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		rk, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("PKCS8 key is not RSA")
		}
		return rk, nil
	default:
		return nil, fmt.Errorf("unsupported PEM type: %s", block.Type)
	}
}
