package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	pb "distributed-cache-system/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	nodeID  string
	address string
	conn    *grpc.ClientConn
	cache   pb.CacheServiceClient
}

type Pool struct {
	mu      sync.RWMutex
	clients map[string]*Client
}

func NewPool() *Pool {
	return &Pool{
		clients: make(map[string]*Client),
	}
}

func (p *Pool) Add(nodeID, address string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.clients[nodeID]; exists {
		return nil
	}

	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to node %s at %s: %w", nodeID, address, err)
	}

	p.clients[nodeID] = &Client{
		nodeID:  nodeID,
		address: address,
		conn:    conn,
		cache:   pb.NewCacheServiceClient(conn),
	}

	return nil
}

func (p *Pool) Get(nodeID, key string) (string, bool, error) {
	client, err := p.getClient(nodeID)
	if err != nil {
		return "", false, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := client.cache.Get(ctx, &pb.GetRequest{Key: key})
	if err != nil {
		return "", false, fmt.Errorf("Get from node %s failed: %w", nodeID, err)
	}

	return resp.Value, resp.Found, nil
}

func (p *Pool) Put(nodeID, key, value string) error {
	client, err := p.getClient(nodeID)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err = client.cache.Put(ctx, &pb.PutRequest{Key: key, Value: value})
	if err != nil {
		return fmt.Errorf("Put to node %s failed: %w", nodeID, err)
	}

	return nil
}

func (p *Pool) Remove(nodeID, key string) error {
	client, err := p.getClient(nodeID)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err = client.cache.Remove(ctx, &pb.RemoveRequest{Key: key})
	if err != nil {
		return fmt.Errorf("Remove from node %s failed: %w", nodeID, err)
	}

	return nil
}

func (p *Pool) RemoveNode(nodeID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if c, exists := p.clients[nodeID]; exists {
		c.conn.Close()
	}
	delete(p.clients, nodeID)
}

func (p *Pool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, c := range p.clients {
		c.conn.Close()
	}
}

func (p *Pool) getClient(nodeID string) (*Client, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	c, exists := p.clients[nodeID]
	if !exists {
		return nil, fmt.Errorf("no connection to node %s", nodeID)
	}
	return c, nil
}

func (p *Pool) GetByAddress(address string) (pb.CacheServiceClient, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, c := range p.clients {
		if c.address == address {
			return c.cache, nil
		}
	}

	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to leader at %s: %w", address, err)
	}
	return pb.NewCacheServiceClient(conn), nil
}
