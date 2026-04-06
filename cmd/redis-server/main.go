package main

import (
	"log"

	"github.com/khalidbm1/build-my-own-redis/internal/server"
)

func main() {

	log.Println("=== Build My Own Redis in Go ===")
	srv := server.New(":6379")

	if err := srv.Start(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
