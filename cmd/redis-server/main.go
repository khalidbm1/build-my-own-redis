package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/khalidbm1/build-my-own-redis/internal/server"
)

func main() {
	log.Println("=== Build Your Own Redis in Go ===")

	srv := server.New(":6379", "dump.aof")

	// Handle graceful shutdown
	// تعامل مع الإيقاف اللطيف
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down...")
		if err := srv.Shutdown(); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
		os.Exit(0)
	}()

	if err := srv.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}

}
