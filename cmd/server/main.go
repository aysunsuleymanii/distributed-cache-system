package main

import (
	"log"

	"distributed-cache-system/internal/cache"
	"distributed-cache-system/internal/server"
)

func main() {
	lru := cache.NewLRUCache[string, string](1000)

	srv := server.New(lru)

	log.Println("starting cache server on :50051")
	if err := srv.Start(":50051"); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
