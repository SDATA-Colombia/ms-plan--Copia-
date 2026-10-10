package grpc

import (
	"fmt"
	"net"

	gogrpc "google.golang.org/grpc"

	"google.golang.org/grpc/reflection"
)

type Server struct {
	server *gogrpc.Server
	port   int
}

func New(port int) *Server {
	s := gogrpc.NewServer()

	// Permite usar grpcurl durante el desarrollo
	reflection.Register(s)

	return &Server{
		server: s,
		port:   port,
	}
}

func (s *Server) Register(register func(*gogrpc.Server)) {
	register(s.server)
}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))

	if err != nil {
		return err
	}

	fmt.Printf("gRPC Server listening on :%d\n", s.port)
	return s.server.Serve(lis)
}

func (s *Server) Stop() {
	s.server.GracefulStop()
}
