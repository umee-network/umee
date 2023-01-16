package query

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	queryTimeout = 15 * time.Second
)

type Client struct {
	GRPCEndpoint string
	grpcConn     *grpc.ClientConn
}

func NewQueryClient(GRPCEndpoint string) (*Client, error) {
	qc := &Client{GRPCEndpoint: GRPCEndpoint}
	err := qc.dialGrpcConn()
	if err != nil {
		return nil, err
	}
	return qc, nil
}

func (qc *Client) dialGrpcConn() (err error) {
	qc.grpcConn, err = grpc.Dial(
		qc.GRPCEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(dialerFunc),
	)
	return err
}
