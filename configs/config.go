package configs

import (
	"os"
	"strings"
)

type Config struct {
	NodeID         string
	NodeAddress    string
	RaftAddress    string
	MetricsAddress string
	Peers          []PeerConfig
	CacheCapacity  uint64
	Bootstrap      bool
	Clean          bool
}

type PeerConfig struct {
	NodeID  string
	Address string
}

func Load() *Config {
	return &Config{
		NodeID:         getEnv("NODE_ID", "node-1"),
		NodeAddress:    getEnv("NODE_ADDRESS", ":50051"),
		RaftAddress:    getEnv("RAFT_ADDRESS", "localhost:50061"),
		MetricsAddress: getEnv("METRICS_ADDRESS", ":9090"),
		Peers:          parsePeers(getEnv("PEERS", "")),
		CacheCapacity:  1000,
		Bootstrap:      getEnv("BOOTSTRAP", "false") == "true",
		Clean:          getEnv("CLEAN", "false") == "true",
	}
}

func parsePeers(raw string) []PeerConfig {
	if raw == "" {
		return []PeerConfig{}
	}

	var peers []PeerConfig
	for _, part := range strings.Split(raw, ",") {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		peers = append(peers, PeerConfig{
			NodeID:  strings.TrimSpace(kv[0]),
			Address: strings.TrimSpace(kv[1]),
		})
	}
	return peers
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
