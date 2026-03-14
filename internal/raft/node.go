package raftnode

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"distributed-cache-system/internal/fsm"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
)

type Node struct {
	raft   *raft.Raft
	fsm    *fsm.CacheFSM
	nodeID string
}

type Config struct {
	NodeID    string
	Address   string
	DataDir   string
	Bootstrap bool
	FSM       *fsm.CacheFSM
}

func NewNode(cfg Config) (*Node, error) {
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data dir: %w", err)
	}

	raftCfg := raft.DefaultConfig()
	raftCfg.LocalID = raft.ServerID(cfg.NodeID)
	raftCfg.HeartbeatTimeout = 500 * time.Millisecond
	raftCfg.ElectionTimeout = 500 * time.Millisecond
	raftCfg.CommitTimeout = 50 * time.Millisecond

	addr, err := raft.NewTCPTransport(cfg.Address, nil, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("failed to create TCP transport: %w", err)
	}

	logStore, err := raftboltdb.NewBoltStore(filepath.Join(cfg.DataDir, "raft-log.db"))
	if err != nil {
		return nil, fmt.Errorf("failed to create log store: %w", err)
	}

	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(cfg.DataDir, "raft-stable.db"))
	if err != nil {
		return nil, fmt.Errorf("failed to create stable store: %w", err)
	}

	snapshotStore, err := raft.NewFileSnapshotStore(cfg.DataDir, 2, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot store: %w", err)
	}

	r, err := raft.NewRaft(raftCfg, cfg.FSM, logStore, stableStore, snapshotStore, addr)
	if err != nil {
		return nil, fmt.Errorf("failed to create raft: %w", err)
	}

	if cfg.Bootstrap {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      raft.ServerID(cfg.NodeID),
					Address: raft.ServerAddress(cfg.Address),
				},
			},
		}
		r.BootstrapCluster(configuration)
	}

	return &Node{
		raft:   r,
		fsm:    cfg.FSM,
		nodeID: cfg.NodeID,
	}, nil
}

func (n *Node) Apply(data []byte) error {
	future := n.raft.Apply(data, 5*time.Second)
	return future.Error()
}

func (n *Node) Join(nodeID, address string) error {
	future := n.raft.AddVoter(
		raft.ServerID(nodeID),
		raft.ServerAddress(address),
		0, 0,
	)
	return future.Error()
}

func (n *Node) IsLeader() bool {
	return n.raft.State() == raft.Leader
}

func (n *Node) LeaderAddress() string {
	addr, _ := n.raft.LeaderWithID()
	return string(addr)
}

func (n *Node) State() string {
	return n.raft.State().String()
}
