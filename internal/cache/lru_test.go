package cache

import "testing"

func TestLRU(t *testing.T) {
	t.Run("Retrieve a value by key", func(t *testing.T) {
		cache := LRUCache[string, int]{
			capacity: 3,
			items: map[string]*Entry[string, int]{
				"Aysun": {key: "Aysun", value: 23},
				"Yavuz": {key: "Yavuz", value: 22},
				"Esin":  {key: "Esin", value: 20},
			},
		}
		expected := 23
		got, ok := cache.Get("Aysun")

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
				123: {key: 123, value: "Nilay"},
				456: {key: 456, value: "Feride"},
			},
		}
		key := 555
		value := "Naser"
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
				123: {key: 123, value: "Nilay"},
				456: {key: 456, value: "Feride"},
			},
		}
		key := 123
		got, ok := cache.Remove(key)
		expected := "Nilay"

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
				123: {key: 123, value: "Nilay"},
				456: {key: 456, value: "Feride"},
			},
		}

		got := cache.Size()
		expected := 2

		if got != expected {
			t.Errorf("expected %d but got %d", expected, got)
		}

	})
	t.Run("Remove everything", func(t *testing.T) {
		cache := LRUCache[int, string]{
			capacity: 2,
			items: map[int]*Entry[int, string]{
				123: {key: 123, value: "Nilay"},
				456: {key: 456, value: "Feride"},
			},
		}
		cache.Clear()
		got := cache.Size()
		expected := 0

		if got != expected {
			t.Errorf("expected %d but got %d", expected, got)
		}
	})
}
