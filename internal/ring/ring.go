package ring

import (
	"crypto/md5"
	"fmt"
	"sort"
	"sync"
)

const defaultReplicas = 150

type Ring struct {
	mu        sync.RWMutex
	replicas  int
	positions []int
	posToNode map[int]string
}

func NewRing() *Ring {
	return &Ring{
		replicas:  defaultReplicas,
		posToNode: make(map[int]string),
	}
}

func (r *Ring) AddNode(nodeID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := 0; i < r.replicas; i++ {
		pos := hash(fmt.Sprintf("%s-%d", nodeID, i))
		r.posToNode[pos] = nodeID
		r.positions = append(r.positions, pos)
	}

	sort.Ints(r.positions)
}

func (r *Ring) RemoveNode(nodeID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := 0; i < r.replicas; i++ {
		pos := hash(fmt.Sprintf("%s-%d", nodeID, i))
		delete(r.posToNode, pos)
	}
	
	r.positions = make([]int, 0, len(r.posToNode))
	for pos := range r.posToNode {
		r.positions = append(r.positions, pos)
	}

	sort.Ints(r.positions)
}

func (r *Ring) GetNode(key string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.positions) == 0 {
		return ""
	}

	pos := hash(key)

	idx := sort.SearchInts(r.positions, pos)

	if idx >= len(r.positions) {
		idx = 0
	}

	return r.posToNode[r.positions[idx]]
}

func hash(key string) int {
	h := md5.Sum([]byte(key))
	return int(h[0])<<24 | int(h[1])<<16 | int(h[2])<<8 | int(h[3])
}
