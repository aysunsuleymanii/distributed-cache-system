package main

import (
	"fmt"
	"log"

	"distributed-cache-system/configs"
	"distributed-cache-system/internal/cache"
	"distributed-cache-system/internal/client"
	"distributed-cache-system/internal/fsm"
	raftnode "distributed-cache-system/internal/raft"
	"distributed-cache-system/internal/ring"
	"distributed-cache-system/internal/server"
)

func main() {
	cfg := configs.Load()
	log.Printf("starting node %s on %s", cfg.NodeID, cfg.NodeAddress)

	lru := cache.NewLRUCache[string, string](cfg.CacheCapacity)

	cacheFSM := fsm.NewCacheFSM(lru)

	raftCfg := raftnode.Config{
		NodeID:    cfg.NodeID,
		Address:   cfg.RaftAddress, // separate port e.g. ":50054"
		DataDir:   fmt.Sprintf("/tmp/raft-%s", cfg.NodeID),
		Bootstrap: cfg.Bootstrap,
		FSM:       cacheFSM,
	}
	rNode, err := raftnode.NewNode(raftCfg)
	if err != nil {
		log.Fatalf("failed to start raft: %v", err)
	}

	r := ring.NewRing()
	for _, peer := range cfg.Peers {
		r.AddNode(peer.NodeID)
	}

	pool := client.NewPool()
	for _, peer := range cfg.Peers {
		if peer.NodeID == cfg.NodeID {
			continue
		}
		if err := pool.Add(peer.NodeID, peer.Address); err != nil {
			log.Printf("warning: could not connect to %s: %v", peer.NodeID, err)
		}
	}

	srv := server.New(lru, r, pool, cfg.NodeID, rNode)

	log.Printf("node %s ready", cfg.NodeID)
	if err := srv.Start(cfg.NodeAddress); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
