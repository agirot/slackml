package helper

import (
	"context"
	"github.com/agirot/slackml/internal/cache"
	"github.com/agirot/slackml/internal/config"
)

const configCtx = "ctx_config"
const cacheCtx = "ctx_cache"

func BuildHydratedContext(ctx context.Context, cfg config.Conf, cache *cache.Disk) context.Context {
	ctx = context.WithValue(ctx, configCtx, &cfg)
	ctx = context.WithValue(ctx, cacheCtx, cache)

	return ctx
}

func GetConfigContext(ctx context.Context) config.Conf {
	val, ok := ctx.Value(configCtx).(*config.Conf)
	if !ok || val == nil {
		panic("config context not found")
	}
	return *val
}

func GetCacheContext(ctx context.Context) *cache.Disk {
	val, ok := ctx.Value(cacheCtx).(*cache.Disk)
	if !ok || val == nil {
		panic("cache context not found")
	}

	return val
}
