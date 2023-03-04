package query

import (
	"context"
	"net"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	queryTimeout = 15 * time.Second
)

type Client struct {
	grpcEndpoint string
	grpcConn     *grpc.ClientConn
}

func NewClient(grpcEndpoint string) (*Client, error) {
	qc := &Client{grpcEndpoint: grpcEndpoint}
	return qc, qc.dialGrpcConn()
}

func (c *Client) dialGrpcConn() (err error) {
	c.grpcConn, err = grpc.Dial(
		c.grpcEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(dialerFunc),
	)
	return err
}

func dialerFunc(_ context.Context, addr string) (net.Conn, error) {
	return Connect(addr)
}

// Connect dials the given address and returns a net.Conn.
// The protoAddr argument should be prefixed with the protocol,
// eg. "tcp://127.0.0.1:8080" or "unix:///tmp/test.sock".
func Connect(protoAddr string) (net.Conn, error) {
	proto, address := protocolAndAddress(protoAddr)
	conn, err := net.Dial(proto, address)
	return conn, err
}

// protocolAndAddress splits an address into the protocol and address components.
// For instance, "tcp://127.0.0.1:8080" will be split into "tcp" and "127.0.0.1:8080".
// If the address has no protocol prefix, the default is "tcp".
func protocolAndAddress(listenAddr string) (string, string) {
	parts := strings.SplitN(listenAddr, "://", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}

	return "tcp", listenAddr
}
