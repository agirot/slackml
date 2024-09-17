package main

import (
	"context"
	"fmt"
	"github.com/agirot/slackml/internal/archive"
	"github.com/agirot/slackml/internal/cache"
	"github.com/agirot/slackml/internal/config"
	"github.com/agirot/slackml/internal/helper"
	"os"
	"time"
)

func main() {
	cfg, err := config.InitConfig()
	if err != nil {
		_, err := fmt.Fprintf(os.Stderr, "%v\n", err)
		if err != nil {
			panic(err)
		}
		os.Exit(1)
	}

	cacheData, err := cache.InitCache(context.Background(), cfg.CacheFile)
	if err != nil {
		_, err := fmt.Fprintf(os.Stderr, "%v\n", err)
		if err != nil {
			panic(err)
		}
		os.Exit(1)
	}

	ctx := helper.BuildHydratedContext(context.Background(), cfg, cacheData)

	if cfg.BackgroundMod {
		archive.BackgroundRun(ctx, 5*time.Second, cfg.RssList)
	} else {
		err := archive.Do(ctx, cfg.RssList)
		if err != nil {
			_, err := fmt.Fprintf(os.Stderr, "%v\n", err)
			if err != nil {
				panic(err)
			}
			os.Exit(1)
		}
	}

	os.Exit(0)
}
