package server

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/khalidbm1/build-my-own-redis/internal/commands"
	"github.com/khalidbm1/build-my-own-redis/internal/protocol"
	"github.com/khalidbm1/build-my-own-redis/internal/storage"
)

type Server struct {
	Addr     string
	store    *storage.Store
	listener net.Listener
}

func New(addr string) *Server {
	return &Server{
		Addr:  addr,
		store: storage.New(),
	}
}

func (s *Server) Start() error {
	// net.Listen creates a TCP listener.
	// "tcp" tells Go we want TCP (not UDP or Unix socket).
	// s.Addr is the address — ":6379" means port 6379 on all interfaces.
	//
	// net.Listen يسوي مستمع TCP.
	// "tcp" يقول إننا نبي TCP (مو UDP أو Unix socket).
	// s.Addr هو العنوان — ":6379" يعني بورت 6379 على كل الواجهات.
	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.Addr, err)
	}
	s.listener = listener
	defer listener.Close()
	log.Printf("Redis server listening on %s\n", s.Addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("failed to accept connection: %w", err)
			continue
		}
		log.Printf("New connection fro %s", conn.RemoteAddr())

		s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	// ALWAYS close the connection when done.
	// If you don't close, connections pile up and you run out of file descriptors.
	//
	// دائماً قفل الاتصال لما تخلص.
	// لو ما تقفل، الاتصالات تتراكم وتخلص الموارد.
	defer conn.Close()

	reader := protocol.NewReader(conn)
	writer := protocol.NewWriter(conn)

	for {
		value, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				log.Printf("Client %s disconnected", conn.RemoteAddr())
				return
			}
			log.Printf("Error reading RESP from %s: %v", conn.RemoteAddr(), err)
			return
		}

		log.Printf("Parsed RESP from %s: Type=%v, Value=%v", conn.RemoteAddr(), value.Type, value)

		response := commands.Dispatch(s.store, value)

		if err := writer.Write(response); err != nil {
			log.Printf("Error writing response to %s: %v", conn.RemoteAddr(), err)
			return
		}
	}
}

func (s *Server) Shutdown() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}
