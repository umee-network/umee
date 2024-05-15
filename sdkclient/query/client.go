package query

import (
	"context"
	"log"
	"net"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	GrpcConn     *grpc.ClientConn
	grpcEndpoint string
	QueryTimeout time.Duration

	logger *log.Logger
}

func NewClient(logger *log.Logger, grpcEndpoint string, queryTimeout time.Duration) (*Client, error) {
	qc := &Client{logger: logger, grpcEndpoint: grpcEndpoint, QueryTimeout: queryTimeout}
	return qc, qc.dialGrpcConn()
}

func (c *Client) dialGrpcConn() (err error) {
	c.GrpcConn, err = grpc.NewClient(
		c.grpcEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(dialerFunc),
	)
	return err
}

func (c Client) NewCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), c.QueryTimeout)
}

func (c Client) NewCtxWithCancel() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
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
