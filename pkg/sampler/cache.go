package sampler

import (
	"sync"
)

const (
	maxCacheSizeBytes = 1 << 30 // 1 GB
)

type (
	cacheEntry struct {
		key  string
		data []float64
		prev *cacheEntry
		next *cacheEntry
	}

	sampleCache struct {
		sync.RWMutex

		entries map[string]*cacheEntry
		head    *cacheEntry
		tail    *cacheEntry
		size    int
	}
)

var (
	cache = sampleCache{
		entries: make(map[string]*cacheEntry),
	}
)

func (c *sampleCache) get(key string) ([]float64, bool) {
	c.RLock()
	defer c.RUnlock()

	entry, ok := c.entries[key]
	if !ok {
		return nil, false
	}

	// move the entry to the front of the list
	if entry.prev != nil {
		entry.prev.next = entry.next
		if entry.next != nil {
			entry.next.prev = entry.prev
		} else {
			c.tail = entry.prev
		}

		entry.prev = nil
		entry.next = c.head
		c.head.prev = entry
		c.head = entry
	}

	return entry.data, true
}

func (c *sampleCache) set(key string, data []float64) {
	c.Lock()
	defer c.Unlock()

	if _, ok := c.entries[key]; ok {
		return
	}

	// remove the oldest entry if the cache is too large
	for c.size+len(data)*8 > maxCacheSizeBytes {
		delete(c.entries, c.tail.key)
		c.size -= len(c.tail.data) * 8

		c.tail = c.tail.prev
		if c.tail != nil {
			c.tail.next = nil
		}
	}

	// add the new entry
	entry := &cacheEntry{
		key:  key,
		data: data,
		next: c.head,
	}
	if c.head != nil {
		c.head.prev = entry
	} else {
		c.tail = entry
	}
	c.head = entry

	c.entries[key] = entry
	c.size += len(data) * 8
}
