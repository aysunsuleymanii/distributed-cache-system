package main

import (
	"log"

	"distributed-cache-system/configs"
	"distributed-cache-system/internal/cache"
	"distributed-cache-system/internal/client"
	"distributed-cache-system/internal/ring"
	"distributed-cache-system/internal/server"
)

func main() {
	// 1. load config from environment variables
	cfg := configs.Load()
	log.Printf("starting node %s on %s", cfg.NodeID, cfg.NodeAddress)

	// 2. create the LRU cache
	lru := cache.NewLRUCache[string, string](cfg.CacheCapacity)

	// 3. create the hash ring and add all peer nodes
	r := ring.NewRing()
	for _, peer := range cfg.Peers {
		r.AddNode(peer.NodeID)
	}

	// 4. create the connection pool and connect to all peers
	pool := client.NewPool()
	for _, peer := range cfg.Peers {
		// don't connect to yourself
		if peer.NodeID == cfg.NodeID {
			continue
		}
		if err := pool.Add(peer.NodeID, peer.Address); err != nil {
			log.Printf("warning: could not connect to peer %s: %v", peer.NodeID, err)
		}
	}

	// 5. create and start the server
	srv := server.New(lru, r, pool, cfg.NodeID)

	log.Printf("node %s ready", cfg.NodeID)
	if err := srv.Start(cfg.NodeAddress); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
