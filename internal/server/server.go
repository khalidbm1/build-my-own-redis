package server

import (
	"fmt"
	"log"
	"net"

	"github.com/khalidbm1/build-my-own-redis/internal/protocol"
)

type Server struct {
	Addr     string
	listener net.Listener
}

func New(addr string) *Server {
	return &Server{
		Addr: addr,
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
	value, err := reader.Read()

	if err != nil {
		log.Printf("Error reading RESP: %v", err)
		return
	}

	log.Printf("Parsed RESP from %s: Type=%v, Value=%v", conn.RemoteAddr(), value.Type, value)

	response := protocol.NewSimpleString("OK")

	if err := writer.Write(response); err != nil{
		log.Printf("Error writing response to: %v", err)
		return
	}

	log.Printf("Sent OK to %s", conn.RemoteAddr())

	// buf := make([]byte, 1024)
	// n, err := conn.Read(buf)

	// if err != nil {
	// 	log.Printf("failed to read from connection: %w", conn.RemoteAddr(), err)
	// 	return
	// }

	// rawRequest := string(buf[:n])
	// log.Printf("Received request from %s:\n%s", conn.RemoteAddr(), rawRequest)
	// response := "+PONNG\r\n"

	// _, err = conn.Write([]byte(response))

	// if err != nil {
	// 	log.Printf("Error Writing response to %s: %v", conn.RemoteAddr(), err)
	// 	return
	// }
	// log.Printf("Sent PONG to %s", conn.RemoteAddr())
}

func (s *Server) Shutdown() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}
