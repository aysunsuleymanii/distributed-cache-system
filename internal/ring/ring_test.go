package ring

import (
	"fmt"
	"testing"
)

func TestGetNodeConsistency(t *testing.T) {
	r := NewRing()
	r.AddNode("node-1")
	r.AddNode("node-2")
	r.AddNode("node-3")

	for i := 0; i < 100; i++ {
		n1 := r.GetNode("user:alice")
		n2 := r.GetNode("user:alice")
		if n1 != n2 {
			t.Fatalf("inconsistent: got %s then %s", n1, n2)
		}
	}
}

func TestAddNodeMovesMinimalKeys(t *testing.T) {
	r := NewRing()
	r.AddNode("node-1")
	r.AddNode("node-2")
	r.AddNode("node-3")

	before := make(map[string]string)
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key-%d", i)
		before[key] = r.GetNode(key)
	}

	r.AddNode("node-4")

	moved := 0
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key-%d", i)
		if r.GetNode(key) != before[key] {
			moved++
		}
	}

	pct := float64(moved) / 1000.0 * 100
	t.Logf("%.1f%% of keys moved when adding node-4", pct)
	if pct < 10 || pct > 40 {
		t.Fatalf("expected ~25%% of keys to move, got %.1f%%", pct)
	}
}

func TestRemoveNodeRedistributes(t *testing.T) {
	r := NewRing()
	r.AddNode("node-1")
	r.AddNode("node-2")
	r.AddNode("node-3")

	r.RemoveNode("node-2")

	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key-%d", i)
		node := r.GetNode(key)
		if node == "node-2" {
			t.Fatalf("key %s still routes to removed node-2", key)
		}
	}
}

func TestEmptyRing(t *testing.T) {
	r := NewRing()
	node := r.GetNode("anykey")
	if node != "" {
		t.Fatalf("expected empty string from empty ring, got %s", node)
	}
}
