package fsm

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"distributed-cache-system/internal/cache"

	"github.com/hashicorp/raft"
)

type CommandType string

const (
	CommandPut    CommandType = "PUT"
	CommandRemove CommandType = "REMOVE"
	CommandClear  CommandType = "CLEAR"
)

type Command struct {
	Type       CommandType `json:"type"`
	Key        string      `json:"key"`
	Value      string      `json:"value"`
	TTLSeconds int64       `json:"ttl_seconds"`
}

type CacheFSM struct {
	cache *cache.LRUCache[string, string]
}

func NewCacheFSM(cache *cache.LRUCache[string, string]) *CacheFSM {
	return &CacheFSM{cache: cache}
}

func (f *CacheFSM) Apply(log *raft.Log) interface{} {
	var cmd Command
	if err := json.Unmarshal(log.Data, &cmd); err != nil {
		return fmt.Errorf("failed to unmarshal command: %w", err)
	}

	switch cmd.Type {
	case CommandPut:
		if cmd.TTLSeconds > 0 {
			ttl := time.Duration(cmd.TTLSeconds) * time.Second
			f.cache.PutWithTTL(cmd.Key, cmd.Value, ttl)
		} else {
			f.cache.Put(cmd.Key, cmd.Value)
		}
	case CommandRemove:
		f.cache.Remove(cmd.Key)
	case CommandClear:
		f.cache.Clear()
	}

	return nil
}

func (f *CacheFSM) Snapshot() (raft.FSMSnapshot, error) {
	items := f.cache.Items()
	return &CacheSnapshot{items: items}, nil
}

func (f *CacheFSM) Restore(snapshot io.ReadCloser) error {
	defer snapshot.Close()

	var items map[string]string
	if err := json.NewDecoder(snapshot).Decode(&items); err != nil {
		return fmt.Errorf("failed to decode snapshot: %w", err)
	}

	f.cache.Clear()
	for k, v := range items {
		f.cache.Put(k, v)
	}
	return nil
}

type CacheSnapshot struct {
	items map[string]string
}

func (s *CacheSnapshot) Persist(sink raft.SnapshotSink) error {
	err := json.NewEncoder(sink).Encode(s.items)
	if err != nil {
		sink.Cancel()
		return err
	}
	return sink.Close()
}

func (s *CacheSnapshot) Release() {}
