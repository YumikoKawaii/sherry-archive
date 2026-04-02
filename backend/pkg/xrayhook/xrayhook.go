// Package xrayhook provides an AWS X-Ray hook for go-redis v9.
// Add it to a redis.Client via client.AddHook(xrayhook.New()).
// Each command becomes a subsegment named "redis.<cmd>" under the active segment.
// Commands on contexts without an active segment are silently skipped
// (requires ContextMissingStrategy = IgnoreError in xray.Configure).
package xrayhook

import (
	"context"

	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/redis/go-redis/v9"
)

type Hook struct{}

func New() *Hook { return &Hook{} }

func (h *Hook) DialHook(next redis.DialHook) redis.DialHook {
	return next
}

func (h *Hook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		_, sub := xray.BeginSubsegment(ctx, "redis."+cmd.Name())
		err := next(ctx, cmd)
		if err != nil && err != redis.Nil {
			sub.AddError(err)
		}
		sub.Close(nil)
		return err
	}
}

func (h *Hook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		_, sub := xray.BeginSubsegment(ctx, "redis.pipeline")
		err := next(ctx, cmds)
		if err != nil {
			sub.AddError(err)
		}
		sub.Close(nil)
		return err
	}
}

// compile-time interface check
var _ redis.Hook = (*Hook)(nil)
