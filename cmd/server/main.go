package main

import (
	"fmt"
	"os"

	"distributed-cache-system/configs"
	"distributed-cache-system/internal/cache"
	"distributed-cache-system/internal/client"
	"distributed-cache-system/internal/fsm"
	"distributed-cache-system/internal/logger"
	"distributed-cache-system/internal/metrics"
	raftnode "distributed-cache-system/internal/raft"
	"distributed-cache-system/internal/ring"
	"distributed-cache-system/internal/server"
	"go.uber.org/zap"
)

func main() {
	cfg := configs.Load()

	// 1. init zap logger
	logger.Init(cfg.NodeID)
	logger.Log.Info("starting node",
		zap.String("node_id", cfg.NodeID),
		zap.String("address", cfg.NodeAddress),
	)

	// 2. clean raft data if CLEAN=true (development only)
	if cfg.Clean {
		dataDir := fmt.Sprintf("/tmp/raft-%s", cfg.NodeID)
		os.RemoveAll(dataDir)
		logger.Log.Info("cleaned raft data",
			zap.String("node_id", cfg.NodeID),
			zap.String("dir", dataDir),
		)
	}

	// 3. start prometheus metrics endpoint
	metrics.Init()
	metrics.StartServer(cfg.MetricsAddress)

	// 4. LRU cache
	lru := cache.NewLRUCache[string, string](cfg.CacheCapacity)

	// 5. FSM — bridges Raft and LRU
	cacheFSM := fsm.NewCacheFSM(lru)

	// 6. Raft node
	raftCfg := raftnode.Config{
		NodeID:    cfg.NodeID,
		Address:   cfg.RaftAddress,
		DataDir:   fmt.Sprintf("/tmp/raft-%s", cfg.NodeID),
		Bootstrap: cfg.Bootstrap,
		FSM:       cacheFSM,
	}
	rNode, err := raftnode.NewNode(raftCfg)
	if err != nil {
		logger.Log.Fatal("failed to start raft", zap.Error(err))
	}

	// 7. Hash ring
	r := ring.NewRing()
	for _, peer := range cfg.Peers {
		r.AddNode(peer.NodeID)
	}

	// 8. Connection pool
	pool := client.NewPool()
	for _, peer := range cfg.Peers {
		if peer.NodeID == cfg.NodeID {
			continue
		}
		if err := pool.Add(peer.NodeID, peer.Address); err != nil {
			logger.Log.Warn("could not connect to peer",
				zap.String("peer", peer.NodeID),
				zap.Error(err),
			)
		}
	}

	// 9. gRPC server
	srv := server.New(lru, r, pool, cfg.NodeID, rNode)

	logger.Log.Info("node ready", zap.String("node_id", cfg.NodeID))
	if err := srv.Start(cfg.NodeAddress); err != nil {
		logger.Log.Fatal("server failed", zap.Error(err))
	}
}
