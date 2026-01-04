package grpc

import (
	"context"
	"fmt"
	"net"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/multimarket-labs/event-pod-services/database"
	"github.com/multimarket-labs/event-pod-services/elasticsearch"
	// TODO: Generate proto files first, then uncomment and import your generated proto package
	// pb "github.com/multimarket-labs/event-pod-services/proto/services"
)

const MaxRecvMessageSize = 1024 * 1024 * 30000

type RpcConfig struct {
	Host string
	Port int
}

type RpcService struct {
	*RpcConfig
	db       *database.DB
	esClient *elasticsearch.ESClient
	// TODO: Embed your generated UnimplementedYourServiceServer here
	// pb.UnimplementedYourServiceServer
	stopped atomic.Bool
}

func NewRpcService(conf *RpcConfig, db *database.DB, esClient *elasticsearch.ESClient) (*RpcService, error) {
	return &RpcService{
		RpcConfig: conf,
		db:        db,
		esClient:  esClient,
	}, nil
}

func (rs *RpcService) Start(ctx context.Context) error {
	go func(rs *RpcService) {
		rpcAddr := fmt.Sprintf("%s:%d", rs.RpcConfig.Host, rs.RpcConfig.Port)
		listener, err := net.Listen("tcp", rpcAddr)
		if err != nil {
			log.Error("Could not start tcp listener", "err", err)
			return
		}

		opt := grpc.MaxRecvMsgSize(MaxRecvMessageSize)

		gs := grpc.NewServer(
			opt,
			grpc.ChainUnaryInterceptor(
				// Add your interceptors here
			),
		)

		reflection.Register(gs)

		// TODO: Register your proto service here
		// pb.RegisterYourServiceServer(gs, rs)

		log.Info("grpc server started", "addr", listener.Addr())

		if err := gs.Serve(listener); err != nil {
			log.Error("start rpc server failed", "err", err)
		}
	}(rs)
	return nil
}

func (rs *RpcService) Stop(ctx context.Context) error {
	rs.stopped.Store(true)
	return nil
}

func (rs *RpcService) Stopped() bool {
	return rs.stopped.Load()
}

// TODO: Implement your gRPC service methods here
// Example:
//
// func (rs *RpcService) HealthCheck(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
//     return &pb.HealthCheckResponse{
//         ReturnCode: pb.ReturnCode_SUCCESS,
//         Message:    "Service is healthy",
//         Version:    "1.0.0",
//         Timestamp:  time.Now().Unix(),
//     }, nil
// }
