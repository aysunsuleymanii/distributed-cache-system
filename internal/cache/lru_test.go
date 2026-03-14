package cache

import "testing"

func TestLRU(t *testing.T) {

	t.Run("Retrieve a value by key", func(t *testing.T) {
		cache := LRUCache[string, int]{
			capacity: 3,
			items: map[string]*Entry[string, int]{
				"Alice":   {key: "Alice", value: 23},
				"Bob":     {key: "Bob", value: 22},
				"Charlie": {key: "Charlie", value: 20},
			},
		}

		expected := 23
		got, ok := cache.Get("Alice")

		if !ok {
			t.Fatalf("expected key does not exist")
		}

		if expected != got {
			t.Errorf("expected %d but got %d", expected, got)
		}
	})

	t.Run("Insert or update a value", func(t *testing.T) {
		cache := LRUCache[int, string]{
			capacity: 2,
			items: map[int]*Entry[int, string]{
				123: {key: 123, value: "Alice"},
				456: {key: 456, value: "Bob"},
			},
		}

		key := 555
		value := "Charlie"

		cache.Put(key, value)

		got, ok := cache.Get(key)

		if !ok {
			t.Fatalf("expected key to exist")
		}

		if got != value {
			t.Errorf("expected %s but got %s", value, got)
		}
	})

	t.Run("Remove a key manually", func(t *testing.T) {
		cache := LRUCache[int, string]{
			capacity: 2,
			items: map[int]*Entry[int, string]{
				123: {key: 123, value: "Alice"},
				456: {key: 456, value: "Bob"},
			},
		}

		key := 123
		expected := "Alice"

		got, ok := cache.Remove(key)

		if !ok {
			t.Fatalf("expected key to exist")
		}

		if got != expected {
			t.Errorf("expected %s but got %s", expected, got)
		}
	})

	t.Run("Return number of elements", func(t *testing.T) {
		cache := LRUCache[int, string]{
			capacity: 2,
			items: map[int]*Entry[int, string]{
				123: {key: 123, value: "Alice"},
				456: {key: 456, value: "Bob"},
			},
		}

		expected := 2
		got := cache.Size()

		if got != expected {
			t.Errorf("expected %d but got %d", expected, got)
		}
	})

	t.Run("Remove everything", func(t *testing.T) {
		cache := LRUCache[int, string]{
			capacity: 2,
			items: map[int]*Entry[int, string]{
				123: {key: 123, value: "Alice"},
				456: {key: 456, value: "Bob"},
			},
		}

		cache.Clear()

		expected := 0
		got := cache.Size()

		if got != expected {
			t.Errorf("expected %d but got %d", expected, got)
		}
	})
}
