package grpc

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client representa un cliente gRPC genérico.
type Client struct {
	Conn *grpc.ClientConn
}

// NewClient establece una conexión gRPC genérica hacia un target.
func NewClient(target string) (*Client, error) {
	conn, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		Conn: conn,
	}, nil
}

// Close cierra la conexión subyacente.
func (c *Client) Close() error {
	if c.Conn != nil {
		return c.Conn.Close()
	}
	return nil
}
