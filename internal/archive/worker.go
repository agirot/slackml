package archive

import (
	"context"
	"errors"
	"fmt"
	"github.com/agirot/slackml/internal/cache"
	"github.com/agirot/slackml/internal/config"
	"github.com/agirot/slackml/internal/helper"
	"github.com/agirot/slackml/internal/slack"
	"github.com/mmcdole/gofeed"
	"os"
	"strings"
	"sync"
	"time"
)

func Do(ctx context.Context, mailingList []config.Rss) error {
	var lastErr error
	wg := sync.WaitGroup{}
	wg.Add(len(mailingList))
	for _, address := range mailingList {
		go func() {
			localErr := readFeed(ctx, address)
			if localErr != nil {
				lastErr = localErr
				_, _ = fmt.Fprintf(os.Stderr, "Error during reading mailing list %s: %v\n", address.ID, localErr)
			}
			wg.Done()
		}()
	}
	wg.Wait()

	return lastErr
}

func BackgroundRun(ctx context.Context, interval time.Duration, mailingList []config.Rss) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		_ = Do(ctx, mailingList)
	}
}

func readFeed(ctx context.Context, mailing config.Rss) error {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(mailing.ID)
	if err != nil {
		return fmt.Errorf("error parsing feed %s: %w", mailing.ID, err)
	}

	var maxItem int
	entry, err := helper.GetCacheContext(ctx).GetLastEntry(ctx, mailing.ID)
	if err == nil {
		//Compare cache datetime to feed
		if feed.PublishedParsed != nil && !feed.PublishedParsed.After(entry.LastFeedUpdated) {
			// No new item to check, return
			return nil
		}
	} else if errors.Is(err, cache.ErrNotFound) {
		maxItem = int(mailing.InitCount)
	}

	var itemSend int
	for _, item := range feed.Items {
		toSend, err := processItem(ctx, *item, mailing.TitleFilters)
		if err != nil {
			return fmt.Errorf("error processing item %s: %w", item.GUID, err)
		}

		if toSend && item.PublishedParsed != nil && feed.UpdatedParsed != nil {
			itemSend++

			err := slack.SendSlack(ctx, item.Published, mailing.ID, item.Title, item.Link)
			if err != nil {
				return fmt.Errorf("error sending slack item %s: %w", item.GUID, err)
			}

			//Most recent only need to be saved
			if itemSend == 1 {
				helper.GetCacheContext(ctx).ReplaceLastEntry(ctx, mailing.ID, cache.Entry{
					LastMailDate:    *item.PublishedParsed,
					LastID:          item.GUID,
					LastFeedUpdated: *feed.UpdatedParsed,
				})
			}
		}

		if itemSend >= maxItem {
			return nil
		}
	}

	if itemSend == 0 && feed.UpdatedParsed != nil {
		helper.GetCacheContext(ctx).RefreshUpdatedAtEntry(ctx, mailing.ID, *feed.UpdatedParsed)
	}

	return nil
}

func processItem(ctx context.Context, item gofeed.Item, filters []string) (bool, error) {
	if len(filters) > 0 {
		found := false
		for _, filter := range filters {
			if strings.Contains(item.Title, filter) {
				found = true
			}
		}
		if !found {
			return false, nil
		}
	}

	return true, nil
}
