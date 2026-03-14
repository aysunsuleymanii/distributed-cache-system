package server

import (
	"context"
	"net"

	"distributed-cache-system/internal/cache"
	"distributed-cache-system/internal/client"
	"distributed-cache-system/internal/ring"
	pb "distributed-cache-system/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type CacheServer struct {
	pb.UnimplementedCacheServiceServer
	nodeID string
	lru    *cache.LRUCache[string, string]
	ring   *ring.Ring
	pool   *client.Pool
}

func New(
	lru *cache.LRUCache[string, string],
	r *ring.Ring,
	pool *client.Pool,
	nodeID string,
) *CacheServer {
	return &CacheServer{
		nodeID: nodeID,
		lru:    lru,
		ring:   r,
		pool:   pool,
	}
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
	owner := s.ring.GetNode(req.Key)

	if owner != s.nodeID {
		err := s.pool.Put(owner, req.Key, req.Value)
		if err != nil {
			return nil, err
		}
		return &pb.PutResponse{}, nil
	}

	s.lru.Put(req.Key, req.Value)
	return &pb.PutResponse{}, nil
}

func (s *CacheServer) Remove(ctx context.Context, req *pb.RemoveRequest) (*pb.RemoveResponse, error) {
	owner := s.ring.GetNode(req.Key)

	if owner != s.nodeID {
		err := s.pool.Remove(owner, req.Key)
		if err != nil {
			return nil, err
		}
		return &pb.RemoveResponse{}, nil
	}

	value, removed := s.lru.Remove(req.Key)
	return &pb.RemoveResponse{Value: value, Removed: removed}, nil
}

func (s *CacheServer) Clear(ctx context.Context, req *pb.ClearRequest) (*pb.ClearResponse, error) {
	s.lru.Clear()
	return &pb.ClearResponse{}, nil
}

func (s *CacheServer) Size(ctx context.Context, req *pb.SizeRequest) (*pb.SizeResponse, error) {
	return &pb.SizeResponse{Size: int64(s.lru.Size())}, nil
}

func (s *CacheServer) Join(ctx context.Context, req *pb.JoinRequest) (*pb.JoinResponse, error) {
	return &pb.JoinResponse{Success: false}, nil
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
