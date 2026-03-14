package server

import (
	"context"
	"net"

	"distributed-cache-system/internal/cache"
	pb "distributed-cache-system/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type CacheServer struct {
	pb.UnimplementedCacheServiceServer
	lru *cache.LRUCache[string, string]
}

func New(lru *cache.LRUCache[string, string]) *CacheServer {
	return &CacheServer{lru: lru}
}

func (s *CacheServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	value, found := s.lru.Get(req.Key)
	return &pb.GetResponse{
		Value: value,
		Found: found,
	}, nil
}

func (s *CacheServer) Put(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error) {
	s.lru.Put(req.Key, req.Value)
	return &pb.PutResponse{}, nil
}

func (s *CacheServer) Remove(ctx context.Context, req *pb.RemoveRequest) (*pb.RemoveResponse, error) {
	value, removed := s.lru.Remove(req.Key)
	return &pb.RemoveResponse{
		Value:   value,
		Removed: removed,
	}, nil
}

func (s *CacheServer) Clear(ctx context.Context, req *pb.ClearRequest) (*pb.ClearResponse, error) {
	s.lru.Clear()
	return &pb.ClearResponse{}, nil
}

func (s *CacheServer) Size(ctx context.Context, req *pb.SizeRequest) (*pb.SizeResponse, error) {
	return &pb.SizeResponse{
		Size: int64(s.lru.Size()),
	}, nil
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
