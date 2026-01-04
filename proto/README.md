# Protocol Buffers Definitions

This directory contains Protocol Buffer (protobuf) definitions for gRPC services.

## Setup

### Install Required Tools

```bash
# Install protoc (Protocol Buffer Compiler)
# macOS
brew install protobuf

# Linux
apt-get install -y protobuf-compiler

# Install Go plugins for protoc
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## Generate Go Code from Proto Files

```bash
# Generate Go code
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/phnxsvc.proto
```

This will generate:
- `proto/services/phnxsvc.pb.go` - Message definitions
- `proto/services/phnxsvc_grpc.pb.go` - gRPC service definitions

## Directory Structure

```
proto/
├── phnxsvc.proto          # Proto definition file
├── services/              # Generated Go code (gitignored)
│   ├── phnxsvc.pb.go
│   └── phnxsvc_grpc.pb.go
└── README.md
```

## Implementing gRPC Services

After generating code, implement the service interface:

```go
package main

import (
    "context"
    pb "your-project/proto/services"
)

type YourServiceServer struct {
    pb.UnimplementedYourServiceServer
}

func (s *YourServiceServer) HealthCheck(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
    return &pb.HealthCheckResponse{
        ReturnCode: pb.ReturnCode_SUCCESS,
        Message:    "Service is healthy",
        Version:    "1.0.0",
        Timestamp:  time.Now().Unix(),
    }, nil
}

func (s *YourServiceServer) StoreEvent(ctx context.Context, req *pb.StoreEventRequest) (*pb.StoreEventResponse, error) {
    // Implement your business logic here
    return &pb.StoreEventResponse{
        ReturnCode: pb.ReturnCode_SUCCESS,
        Message:    "Event stored successfully",
        EventId:    "evt_123",
    }, nil
}
```

## Resources

- [Protocol Buffers Documentation](https://protobuf.dev/)
- [gRPC Go Quick Start](https://grpc.io/docs/languages/go/quickstart/)
- [Protocol Buffer Language Guide](https://protobuf.dev/programming-guides/proto3/)
