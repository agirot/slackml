package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type Disk struct {
	sync.Mutex
	Data      data
	CachePath string
}

type Entry struct {
	LastMailDate    time.Time
	LastID          string
	LastFeedUpdated time.Time
}

type data map[string]Entry

type Item string

var ErrNotFound = errors.New("cache entry not found")

func InitCache(ctx context.Context, cachePath string) (*Disk, error) {
	file, err := os.OpenFile(cachePath, os.O_RDWR|os.O_CREATE, 0666)
	defer file.Close()
	if err != nil {
		return &Disk{}, fmt.Errorf("failed open cache file: %w", err)
	}

	decoder := json.NewDecoder(file)
	var data map[string]Entry
	err = decoder.Decode(&data)
	if err != nil {
		log.Println("failed decoding cache file, create a new one")
		data = make(map[string]Entry)
	}

	disk := Disk{
		CachePath: cachePath,
		Data:      data,
	}
	return &disk, nil
}

func (disk *Disk) writeCache(ctx context.Context) {
	b, err := json.Marshal(disk.Data)
	if err != nil {
		log.Println("failed encoding cache file, create a new one")
	}

	err = os.WriteFile(disk.CachePath, b, 0600)
	if err != nil {
		log.Println("failed writing cache file, create a new one")
	}
}

func (disk *Disk) GetLastEntry(ctx context.Context, id string) (Entry, error) {
	disk.Lock()
	defer disk.Unlock()

	if entry, ok := disk.Data[id]; ok {
		return entry, nil
	}
	return Entry{}, ErrNotFound
}

func (disk *Disk) ReplaceLastEntry(ctx context.Context, id string, item Entry) {
	disk.Lock()
	defer disk.Unlock()

	delete(disk.Data, id)
	disk.Data[id] = item
	disk.writeCache(ctx)
}

func (disk *Disk) RefreshUpdatedAtEntry(ctx context.Context, id string, lastFeedUpdated time.Time) {
	disk.Lock()
	defer disk.Unlock()

	if entry, ok := disk.Data[id]; ok {
		entry.LastFeedUpdated = lastFeedUpdated
		disk.Data[id] = entry
	}
}
