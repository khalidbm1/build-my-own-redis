package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/khalidbm1/build-my-own-redis/internal/commands"
	"github.com/khalidbm1/build-my-own-redis/internal/persistence"
	"github.com/khalidbm1/build-my-own-redis/internal/protocol"
	"github.com/khalidbm1/build-my-own-redis/internal/storage"
)

type Server struct {
	Addr     string
	store    *storage.Store
	aof      *persistence.AOF
	listener net.Listener

	activeConnections sync.WaitGroup
	shutdownChan      chan bool
}

func New(addr string, aofFilename string) *Server {
	store := storage.New()
	return &Server{
		Addr:         addr,
		store:        store,
		aof:          persistence.NewAOF(aofFilename),
		shutdownChan: make(chan bool),
	}
}

func (s *Server) Start() error {
	// Replay existing AOF before accepting writes.
	if err := s.aof.Load(s.store); err != nil {
		log.Printf("Warning: Failed to replay AOF: %v", err)
	}
	if err := s.aof.Open(); err != nil {
		log.Printf("Warning: Failed to open AOF for append: %v", err)
	}

	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.Addr, err)
	}

	s.listener = listener
	defer listener.Close()
	log.Printf("Redis server listening on %s\n", s.Addr)
	// Accept incoming connections loop
	for {
		select {
		case <-s.shutdownChan:
			return nil
		default:
		}

		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-s.shutdownChan:
				return nil
			default:
				log.Printf("Error accepting connection: %v", err)
				continue
			}
		}
		log.Printf("New Connection from %s", conn.RemoteAddr())

		s.activeConnections.Add(1)
		go s.handleConnection(conn)

	}
}

func (s *Server) handleConnection(conn net.Conn) {
	// activeConnections done
	defer s.activeConnections.Done()
	// connection close
	defer conn.Close()

	reader := protocol.NewReader(conn)
	writer := protocol.NewWriter(conn)

	for {
		value, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				log.Printf("Client %s disconnected", conn.RemoteAddr())
			} else {
				log.Printf("Read error from %s: %v", conn.RemoteAddr(), err)
			}
			return
		}

		log.Printf("[%s] Received command: %v", conn.RemoteAddr(), value)

		if persistence.IsWriteCommand(value) {
			if err := s.aof.Append(value); err != nil {
				log.Printf("Error appending to AOF: %v", err)
			}
		}

		response := commands.Dispatch(s.store, value)
		if err := writer.Write(response); err != nil {
			log.Printf("Error writing response: %v", err)
			return
		}
	}
}

func (s *Server) Shutdown() error {
	log.Println("Starting graceful shutdown...")
	close(s.shutdownChan)

	var firstErr error
	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			firstErr = err
		}
	}

	log.Println("Waiting for active connections to finish...")
	s.activeConnections.Wait()

	if err := s.aof.Close(); err != nil && firstErr == nil {
		firstErr = err
	}

	s.store.Close()

	log.Println("Server shut down complete.")
	return firstErr
}
