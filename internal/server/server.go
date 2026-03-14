package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	"distributed-cache-system/internal/cache"
	"distributed-cache-system/internal/client"
	"distributed-cache-system/internal/fsm"
	raftnode "distributed-cache-system/internal/raft"
	"distributed-cache-system/internal/ring"
	pb "distributed-cache-system/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type CacheServer struct {
	pb.UnimplementedCacheServiceServer
	nodeID   string
	lru      *cache.LRUCache[string, string]
	ring     *ring.Ring
	pool     *client.Pool
	raftNode *raftnode.Node // added
}

func New(
	lru *cache.LRUCache[string, string],
	r *ring.Ring,
	pool *client.Pool,
	nodeID string,
	raftNode *raftnode.Node,
) *CacheServer {
	return &CacheServer{
		nodeID:   nodeID,
		lru:      lru,
		ring:     r,
		pool:     pool,
		raftNode: raftNode,
	}
}

func (s *CacheServer) applyCommand(cmd fsm.Command) error {
	if s.raftNode.IsLeader() {
		data, err := json.Marshal(cmd)
		if err != nil {
			return err
		}
		return s.raftNode.Apply(data)
	}

	leaderAddr := s.raftNode.LeaderAddress()
	if leaderAddr == "" {
		return fmt.Errorf("no leader elected yet, try again")
	}

	ctx := context.Background()
	leaderClient, err := s.pool.GetByAddress(leaderAddr)
	if err != nil {
		return fmt.Errorf("cannot reach leader at %s: %w", leaderAddr, err)
	}

	switch cmd.Type {
	case fsm.CommandPut:
		_, err = leaderClient.Put(ctx, &pb.PutRequest{Key: cmd.Key, Value: cmd.Value})
	case fsm.CommandRemove:
		_, err = leaderClient.Remove(ctx, &pb.RemoveRequest{Key: cmd.Key})
	case fsm.CommandClear:
		_, err = leaderClient.Clear(ctx, &pb.ClearRequest{})
	}
	return err
}

func (s *CacheServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	owner := s.ring.GetNode(req.Key)
	if owner != s.nodeID {
		value, found, err := s.pool.Get(owner, req.Key)
		if err != nil {
			return nil, err
		}
		return &pb.GetResponse{Value: value, Found: found}, nil
	}

	value, found := s.lru.Get(req.Key)
	return &pb.GetResponse{Value: value, Found: found}, nil
}

func (s *CacheServer) Put(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error) {
	err := s.applyCommand(fsm.Command{
		Type:  fsm.CommandPut,
		Key:   req.Key,
		Value: req.Value,
	})
	if err != nil {
		return nil, err
	}
	return &pb.PutResponse{}, nil
}

func (s *CacheServer) Remove(ctx context.Context, req *pb.RemoveRequest) (*pb.RemoveResponse, error) {
	err := s.applyCommand(fsm.Command{
		Type: fsm.CommandRemove,
		Key:  req.Key,
	})
	if err != nil {
		return nil, err
	}
	return &pb.RemoveResponse{}, nil
}

func (s *CacheServer) Clear(ctx context.Context, req *pb.ClearRequest) (*pb.ClearResponse, error) {
	err := s.applyCommand(fsm.Command{Type: fsm.CommandClear})
	if err != nil {
		return nil, err
	}
	return &pb.ClearResponse{}, nil
}

func (s *CacheServer) Size(ctx context.Context, req *pb.SizeRequest) (*pb.SizeResponse, error) {
	return &pb.SizeResponse{Size: int64(s.lru.Size())}, nil
}

func (s *CacheServer) Join(ctx context.Context, req *pb.JoinRequest) (*pb.JoinResponse, error) {
	if !s.raftNode.IsLeader() {
		return &pb.JoinResponse{
			Success:       false,
			LeaderAddress: s.raftNode.LeaderAddress(),
		}, nil
	}

	err := s.raftNode.Join(req.NodeId, req.Address)
	if err != nil {
		return &pb.JoinResponse{Success: false}, err
	}
	return &pb.JoinResponse{Success: true}, nil
}

func (s *CacheServer) Start(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	pb.RegisterCacheServiceServer(grpcServer, s)
	reflection.Register(grpcServer)

	return grpcServer.Serve(listener)
}
