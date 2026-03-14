package cache

import (
	"sync"
	"time"
)

type Entry[K comparable, V any] struct {
	key       K
	value     V
	expiresAt time.Time
	prev      *Entry[K, V]
	next      *Entry[K, V]
}

type LRUCache[K comparable, V any] struct {
	capacity uint64
	items    map[K]*Entry[K, V]
	head     *Entry[K, V]
	tail     *Entry[K, V]
	mu       sync.RWMutex
}

func NewLRUCache[K comparable, V any](capacity uint64) *LRUCache[K, V] {
	c := &LRUCache[K, V]{
		capacity: capacity,
		items:    make(map[K]*Entry[K, V]),
	}
	go c.sweep()
	return c
}

func (cache *LRUCache[K, V]) Get(key K) (value V, exists bool) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	entry, exists := cache.items[key]
	if !exists {
		return value, false
	}

	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		cache.removeNode(entry)
		delete(cache.items, key)
		return value, false
	}

	cache.moveToHead(entry)
	return entry.value, true
}

func (cache *LRUCache[K, V]) Put(key K, value V) {
	cache.PutWithTTL(key, value, 0)
}

func (cache *LRUCache[K, V]) PutWithTTL(key K, value V, ttl time.Duration) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	if entry, exists := cache.items[key]; exists {
		entry.value = value
		entry.expiresAt = expiresAt
		cache.moveToHead(entry)
		return
	}

	entry := &Entry[K, V]{
		key:       key,
		value:     value,
		expiresAt: expiresAt,
	}

	cache.items[key] = entry
	cache.addToHead(entry)

	if uint64(len(cache.items)) > cache.capacity {
		cache.removeTail()
	}
}

func (cache *LRUCache[K, V]) sweep() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		cache.mu.Lock()
		now := time.Now()
		for key, entry := range cache.items {
			if !entry.expiresAt.IsZero() && now.After(entry.expiresAt) {
				cache.removeNode(entry)
				delete(cache.items, key)
			}
		}
		cache.mu.Unlock()
	}
}

func (cache *LRUCache[K, V]) Remove(key K) (value V, successful bool) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	entry, ok := cache.items[key]
	if !ok {
		return value, false
	}

	cache.removeNode(entry)
	delete(cache.items, key)
	return entry.value, true
}

func (cache *LRUCache[K, V]) Size() int {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	return len(cache.items)
}

func (cache *LRUCache[K, V]) Clear() {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.head = nil
	cache.tail = nil
	cache.items = make(map[K]*Entry[K, V])
}

func (cache *LRUCache[K, V]) Items() map[K]V {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	result := make(map[K]V, len(cache.items))
	for k, v := range cache.items {
		result[k] = v.value
	}
	return result
}

func (cache *LRUCache[K, V]) addToHead(node *Entry[K, V]) {
	node.prev = nil
	node.next = cache.head
	if cache.head != nil {
		cache.head.prev = node
	}
	cache.head = node
	if cache.tail == nil {
		cache.tail = node
	}
}

func (cache *LRUCache[K, V]) removeNode(node *Entry[K, V]) {
	if node.prev != nil {
		node.prev.next = node.next
	} else {
		cache.head = node.next
	}
	if node.next != nil {
		node.next.prev = node.prev
	} else {
		cache.tail = node.prev
	}
	node.prev = nil
	node.next = nil
}

func (cache *LRUCache[K, V]) removeTail() *Entry[K, V] {
	node := cache.tail
	if node == nil {
		return nil
	}
	cache.removeNode(node)
	delete(cache.items, node.key)
	return node
}

func (cache *LRUCache[K, V]) moveToHead(node *Entry[K, V]) {
	cache.removeNode(node)
	cache.addToHead(node)
}
