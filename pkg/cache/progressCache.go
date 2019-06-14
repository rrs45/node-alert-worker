package cache

import (
	"sync"
	"time"

	"github.com/box-node-alert-worker/pkg/types"
	log "github.com/sirupsen/logrus"
)

//StatusCache is a struct to store InProgress of remediation
type StatusCache struct {
	Items               map[string]types.Status
	CacheExpireInterval time.Duration
	Locker              *sync.RWMutex
}

//NewStatusCache instantiates and returns a new cache
func NewStatusCache(cacheExpireInterval string) *StatusCache {
	interval, _ := time.ParseDuration(cacheExpireInterval)
	return &StatusCache{
		Items:               make(map[string]types.Status),
		CacheExpireInterval: interval,
		Locker:              new(sync.RWMutex),
	}
}

//PurgeExpired expires cache items older than specified purge interval
func (cache *StatusCache) PurgeExpired() {
	ticker := time.NewTicker(cache.CacheExpireInterval)
	for {
		select {
		case <-ticker.C:
			log.Info("CacheManager - Attempting to delete expired entries")
			cache.Locker.Lock()
			for cond, result := range cache.Items {
				if time.Since(result.Timestamp) > cache.CacheExpireInterval {
					log.Info("CacheManager - Deleting expired entry for ", cond)
					delete(cache.Items, cond)
				}
			}
			cache.Locker.Unlock()

		}
	}
}

//Set appends entry to the slice
func (cache *StatusCache) Set(key string, action types.Status) {
	cache.Locker.Lock()
	defer cache.Locker.Unlock()
	cache.Items[key] = action	
}

//GetAll returns current entries of a cache
func (cache *StatusCache) GetAll() map[string]types.Status {
	cache.Locker.RLock()
	defer cache.Locker.RUnlock()
	return cache.Items
}

//Count returns number of items in cache
func (cache *StatusCache) Count() int {
	cache.Locker.RLock()
	defer cache.Locker.RUnlock()
	return len(cache.Items)
}

//GetItem returns value of a given key and whether it exist or not
func (cache *StatusCache) GetItem(key string) (types.Status, bool) {
	cache.Locker.RLock()
	defer cache.Locker.RUnlock()
	val, found := cache.Items[key]
	if found {
		return val, true
	}
	return types.Status{}, false
}

//DelItem deletes a cache item with a given key
func (cache *StatusCache) DelItem(key string)  {
	cache.Locker.Lock()
	delete(cache.Items,key)
	cache.Locker.Unlock()
}


