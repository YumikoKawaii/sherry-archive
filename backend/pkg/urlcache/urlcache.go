package urlcache

import (
	"context"
	"errors"
	"net/url"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const keyPrefix = "{presign}:"

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

// ResolveMany resolves presigned URLs for multiple keys efficiently:
// one MGET fetches all cached values in a single Redis round trip,
// then only cache misses are signed — in parallel.
// The returned slice is ordered the same as keys.
func (c *URLCache) ResolveMany(ctx context.Context, keys []string) ([]string, error) {
	if len(keys) == 0 {
		return nil, nil
	}

	cacheKeys := make([]string, len(keys))
	for i, k := range keys {
		cacheKeys[i] = keyPrefix + k
	}

	// Single round trip to Redis for all keys.
	vals, err := c.rdb.MGet(ctx, cacheKeys...).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		// Redis unavailable — fall back to signing all keys sequentially.
		vals = make([]any, len(keys))
	}

	urls := make([]string, len(keys))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i, v := range vals {
		if s, ok := v.(string); ok && s != "" {
			urls[i] = s
			continue
		}
		if keys[i] == "" {
			continue
		}
		// Cache miss — sign in parallel.
		wg.Add(1)
		go func(idx int, key string) {
			defer wg.Done()
			u, err := c.signer.PresignedGetURL(ctx, key)
			if err != nil {
				return
			}
			raw := u.String()
			mu.Lock()
			urls[idx] = raw
			mu.Unlock()
			c.rdb.Set(ctx, keyPrefix+key, raw, c.ttl) // best-effort
		}(i, keys[i])
	}

	wg.Wait()
	return urls, nil
}
