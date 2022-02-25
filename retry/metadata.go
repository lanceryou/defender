package retry

import (
	"context"

	"github.com/lanceryou/defender/pkg/metadata"
)

// RetryContext retry context
type RetryContext interface {
	CanRetry(ctx context.Context) bool
}

type RetryContextFunc func(ctx context.Context) bool

func (r RetryContextFunc) CanRetry(ctx context.Context) bool {
	return r(ctx)
}

// NopCanRetry
func NopCanRetry(ctx context.Context) bool {
	return true
}

// NoRetryMetadataContext
func NoRetryMetadataContext(ctx context.Context) context.Context {
	return RetryMetadataFromContext(ctx, "no_retry", "1")
}

// IsNoRetry
func IsNoRetry(ctx context.Context) bool {
	return MetadataContains(ctx, "no_retry", "1")
}

// RetryMetadataFromContext
func RetryMetadataFromContext(ctx context.Context, kv ...string) context.Context {
	md := metadata.FromContext(ctx)
	if md == nil {
		return metadata.NewMetadataFromContext(ctx, metadata.Pairs(kv...))
	}

	for k, v := range metadata.Pairs(kv...) {
		md.Set(k, v)
	}

	return ctx
}

// MetadataContains
func MetadataContains(ctx context.Context, kv ...string) bool {
	md := metadata.FromContext(ctx)
	if md == nil {
		return false
	}

	for k, v := range metadata.Pairs(kv...) {
		if md.Get(k) == v {
			return true
		}
	}

	return false
}

// MetadataContainsAll
func MetadataContainsAll(ctx context.Context, kv ...string) bool {
	md := metadata.FromContext(ctx)
	if md == nil {
		return false
	}

	for k, v := range metadata.Pairs(kv...) {
		if md.Get(k) != v {
			return false
		}
	}

	return true
}
