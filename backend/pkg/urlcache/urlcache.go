package urlcache

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/redis/go-redis/v9"
)

const keyPrefix = "presign:"

// Signer generates signed/presigned URLs for object storage keys.
type Signer interface {
	PresignedGetURL(ctx context.Context, key string) (*url.URL, error)
}

// URLCache caches presigned URLs in Redis so the same URL is reused
// within the presign window. All callers sharing the same key get the same
// URL, which is a prerequisite for CDN caching.
type URLCache struct {
	signer Signer
	rdb    *redis.Client
	ttl    time.Duration // slightly less than presign expiry
}

// New creates a URLCache. ttl is set to presignExpiry minus a 5-minute buffer
// so cached URLs are always comfortably valid when served.
func New(s Signer, rdb *redis.Client, presignExpiry time.Duration) *URLCache {
	ttl := presignExpiry - 5*time.Minute
	if ttl <= 0 {
		ttl = presignExpiry
	}
	return &URLCache{signer: s, rdb: rdb, ttl: ttl}
}

// Resolve returns a presigned URL for the given object key, using Redis as a
// cache-aside. On a Redis error the cache is bypassed and a fresh URL is
// generated so the request never fails due to cache unavailability.
func (c *URLCache) Resolve(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", nil
	}

	cacheKey := keyPrefix + key

	if cached, err := c.rdb.Get(ctx, cacheKey).Result(); err == nil {
		return cached, nil
	} else if !errors.Is(err, redis.Nil) {
		// Redis unavailable — fall through, don't fail the request
	}

	u, err := c.signer.PresignedGetURL(ctx, key)
	if err != nil {
		return "", err
	}
	raw := u.String()

	c.rdb.Set(ctx, cacheKey, raw, c.ttl) // best-effort, ignore error

	return raw, nil
}
