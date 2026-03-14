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
