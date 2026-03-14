package cache

type Entry[K comparable, V any] struct {
	key   K
	value V
	prev  *Entry[K, V]
	next  *Entry[K, V]
}

type LRUCache[K comparable, V any] struct {
	capacity uint64
	items    map[K]*Entry[K, V]
	head     *Entry[K, V]
	tail     *Entry[K, V]
}

func NewLRUCache[K comparable, V any](capacity uint64) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		capacity: capacity,
		items:    make(map[K]*Entry[K, V]),
	}
}

func (cache *LRUCache[K, V]) Get(key K) (value V, exists bool) {
	entry, exists := cache.items[key]
	if !exists {
		return value, false
	}
	return entry.value, true
}

func (cache *LRUCache[K, V]) Put(key K, value V) {
	entry := &Entry[K, V]{
		key:   key,
		value: value,
	}

	cache.items[key] = entry
}

func (cache *LRUCache[K, V]) Remove(key K) (value V, successful bool) {
	entry, ok := cache.items[key]
	if !ok {
		return value, false
	}

	value = entry.value
	delete(cache.items, key)

	return value, true
}

func (cache *LRUCache[K, V]) Size() int {
	return len(cache.items)
}

func (cache *LRUCache[K, V]) Clear() {
	cache.head = nil
	cache.tail = nil
	cache.items = make(map[K]*Entry[K, V])
}
