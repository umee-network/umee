package query

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	queryTimeout = 15 * time.Second
)

type QueryClient struct {
	GRPCEndpoint string
	grpcConn     *grpc.ClientConn
}

func NewQueryClient(GRPCEndpoint string) *QueryClient {
	qc := &QueryClient{GRPCEndpoint: GRPCEndpoint}
	qc.dialGrpcConn()
	return qc
}

func (qc *QueryClient) dialGrpcConn() (err error) {
	qc.grpcConn, err = grpc.Dial(
		qc.GRPCEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(dialerFunc),
	)
	return err
}
