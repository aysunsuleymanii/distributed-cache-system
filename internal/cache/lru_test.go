package cache

import "testing"

func TestLRU(t *testing.T) {

	t.Run("Retrieve a value by key", func(t *testing.T) {
		cache := NewLRUCache[string, int](3)

		cache.Put("Alice", 23)
		cache.Put("Bob", 22)
		cache.Put("Charlie", 20)

		expected := 23
		got, ok := cache.Get("Alice")

		if !ok {
			t.Fatalf("expected key to exist")
		}

		if expected != got {
			t.Errorf("expected %d but got %d", expected, got)
		}
	})

	t.Run("Insert or update a value", func(t *testing.T) {
		cache := NewLRUCache[int, string](3)

		cache.Put(123, "Alice")
		cache.Put(456, "Bob")

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
		cache := NewLRUCache[int, string](3)

		cache.Put(123, "Alice")
		cache.Put(456, "Bob")

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
		cache := NewLRUCache[int, string](3)

		cache.Put(123, "Alice")
		cache.Put(456, "Bob")

		expected := 2
		got := cache.Size()

		if got != expected {
			t.Errorf("expected %d but got %d", expected, got)
		}
	})

	t.Run("Remove everything", func(t *testing.T) {
		cache := NewLRUCache[int, string](3)

		cache.Put(123, "Alice")
		cache.Put(456, "Bob")

		cache.Clear()

		expected := 0
		got := cache.Size()

		if got != expected {
			t.Errorf("expected %d but got %d", expected, got)
		}
	})

	t.Run("Evicts least recently used item when capacity exceeded", func(t *testing.T) {
		cache := NewLRUCache[int, string](2)

		cache.Put(1, "A")
		cache.Put(2, "B")

		cache.Get(1)

		cache.Put(3, "C")

		_, ok := cache.Get(2)
		if ok {
			t.Fatalf("expected key 2 to be evicted")
		}
	})
}
