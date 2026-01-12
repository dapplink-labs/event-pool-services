# Event Pod Services

A foundational Go project for building blockchain services and event-driven applications.

## Features

### Core Infrastructure
- **Complete Service Framework**: HTTP API, gRPC, WebSocket support
- **Database Layer**: PostgreSQL with GORM ORM, transaction support, SQL migrations
- **Caching**: Redis integration with connection pooling, distributed locks, rate limiting
- **Search**: Elasticsearch integration for full-text search and analytics
- **Monitoring**: Prometheus metrics collection

### Blockchain Integration
- **Ethereum Client**: Full go-ethereum integration
- **Smart Contract Bindings**: Tools and patterns for generating Go bindings from ABIs
- **Transaction Relayer**: Framework for relaying transactions to blockchain networks
- **Transaction Manager**: Nonce management, gas estimation, retry logic
- **Event Indexing**: Listen to and process blockchain events

### Utilities
- **Common Library**: Rich set of utility functions
  - JWT authentication
  - Encryption/hashing
  - Retry strategies
  - Time utilities
  - String manipulation
  - TOTP generation
- **Third-party Integrations**:
  - Email (SMTP)
  - SMS (Alibaba Cloud)
  - Object Storage (AWS S3, Qiniu Kodo, MinIO)
  - Wallet signature verification (SIWE)

## Directory Structure

```
event-pod-services/
├── abis/                       # Smart contract ABI files
├── bindings/                   # Generated Go bindings for contracts
├── cache/                      # Redis client wrapper
├── cmd/                        # Application entrypoints
│   └── event-services/         # Main application
├── common/                     # Common utilities
│   ├── bigint/                 # Big integer handling
│   ├── clock/                  # Time utilities
│   ├── httputil/               # HTTP server utilities
│   ├── json2/                  # JSON utilities
│   ├── retry/                  # Retry strategies
│   ├── utils/                  # JWT, encryption, etc.
│   └── ...
├── config/                     # Configuration management
├── database/                   # Database layer
│   └── backend/                # Data models
├── elasticsearch/              # Elasticsearch client
├── metrics/                    # Prometheus metrics
├── migrations/                 # SQL migration files
├── proto/                      # Protocol Buffer definitions
├── relayer/                    # Transaction relayer
│   ├── driver/                 # Blockchain interaction driver
│   └── txmgr/                  # Transaction manager
├── services/                   # Service layer
│   ├── api/                    # HTTP API service
│   ├── common/                 # Shared service components
│   ├── gprc/                   # gRPC service
│   └── websocket/              # WebSocket service
├── worker/                     # Background task workers
├── config.example.yaml         # Example configuration file
└── README.md
```

## Getting Started

### Prerequisites

- Go 1.24 or higher
- PostgreSQL 12 or higher
- (Optional) Redis
- (Optional) Elasticsearch
- (Optional) Ethereum node access (Infura, Alchemy, or your own node)

### Installation

1. Clone the repository:
```bash
git clone https://github.com/multimarket-labs/event-pod-services.git
cd event-pod-services
```

2. Install dependencies:
```bash
go mod download
```

3. Copy and configure the example config:
```bash
cp config.example.yaml config.yaml
# Edit config.yaml with your database credentials and other settings
```

4. Run database migrations:
```bash
go build -o event-services ./cmd/event-services
./event-services migrate -c config.yaml
```

### Running the Services

The application supports multiple modes:

#### 1. HTTP API Server
```bash
./event-services api -c config.yaml
```
Starts the HTTP API server on the configured port (default: 8080).

#### 2. gRPC Server
```bash
./event-services rpc -c config.yaml
```
Starts the gRPC server on the configured port (default: 9090).

#### 3. Event Indexer (Blockchain Node)
```bash
./event-services index -c config.yaml
```
Starts the blockchain event indexer that listens to contract events.

#### 4. Database Migration
```bash
./event-services migrate -c config.yaml
```
Runs database migrations from the `migrations/` directory.

## Development Guide

### Adding Smart Contract Integration

1. **Place ABI file**:
```bash
# Put your contract ABI in abis/ directory
mkdir -p abis/MyContract.sol
cp MyContract.json abis/MyContract.sol/
```

2. **Generate Go bindings**:
```bash
# Install abigen
go install github.com/ethereum/go-ethereum/cmd/abigen@latest

# Generate bindings
abigen --abi abis/MyContract.sol/MyContract.json \
       --pkg bindings \
       --type MyContract \
       --out bindings/my_contract.go
```

3. **Use in your code**:
```go
import (
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/multimarket-labs/event-pod-services/bindings"
)

client, _ := ethclient.Dial("https://mainnet.infura.io/v3/YOUR_KEY")
contract, _ := bindings.NewMyContract(contractAddress, client)
```

### Adding Database Models

1. **Create model** in `database/backend/`:
```go
package backend

type User struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    Email     string    `gorm:"uniqueIndex;not null" json:"email"`
    Username  string    `gorm:"index" json:"username"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

func (User) TableName() string {
    return "users"
}
```

2. **Add migration** in `migrations/`:
```sql
CREATE TABLE IF NOT EXISTS users(
    guid        TEXT PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    email       VARCHAR NOT NULL UNIQUE,
    username    VARCHAR NOT NULL,
    created     INTEGER CHECK (created > 0),
    updated     INTEGER CHECK (updated > 0)
);
```

3. **Create DB interface**:
```go
type UserDB interface {
    Create(user *User) error
    FindByID(id uint) (*User, error)
    Update(user *User) error
    Delete(id uint) error
}
```

### Adding API Endpoints

1. **Create route handler** in `services/api/routes/`:
```go
func (rs *Routes) HandleGetUser(w http.ResponseWriter, r *http.Request) {
    userID := chi.URLParam(r, "id")
    // Your logic here
}
```

2. **Register route** in `services/api/routes/routes.go`:
```go
r.Get("/users/{id}", rs.HandleGetUser)
```

### Adding gRPC Services

1. **Define proto** in `proto/`:
```protobuf
service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse) {}
}
```

2. **Generate code**:
```bash
protoc --go_out=. --go-grpc_out=. proto/your_service.proto
```

3. **Implement service** in `services/gprc/`:
```go
type UserServiceServer struct {
    pb.UnimplementedUserServiceServer
}

func (s *UserServiceServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
    // Your implementation
}
```

## Configuration

See `config.example.yaml` for all available configuration options.

Key configuration sections:
- `master_db`: PostgreSQL database connection
- `redis`: Redis cache configuration
- `elasticsearch_config`: Elasticsearch settings
- `rpcs`: Blockchain RPC endpoints and contract addresses
- `email_config`, `sms_config`: Third-party service integrations
- `s3_config`, `kodo_config`, `minio_config`: Object storage

## Architecture

### Service Layers

1. **API Layer** (`services/api/`): HTTP REST API endpoints
2. **gRPC Layer** (`services/gprc/`): gRPC service implementations
3. **WebSocket Layer** (`services/websocket/`): Real-time WebSocket connections
4. **Service Layer** (`services/api/service/`): Business logic
5. **Database Layer** (`database/`): Data access and models
6. **Relayer Layer** (`relayer/`): Blockchain transaction relaying

### Blockchain Components

- **Driver** (`relayer/driver/`): Blockchain interaction engine
- **Transaction Manager** (`relayer/txmgr/`): Transaction lifecycle management
- **Worker** (`worker/`): Background job processing

## Monitoring

The service exposes Prometheus metrics on the configured metrics port (default: 7300).

Access metrics at: `http://localhost:7300/metrics`

## Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for a specific package
go test ./database/...
```

## Building for Production

```bash
# Build binary
go build -o event-services ./cmd/event-services

# Build with optimizations
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o event-services ./cmd/event-services
```

## Docker Deployment

```bash
# Build Docker image
docker build -t event-pod-services:latest .

# Run with Docker Compose
docker-compose up -d
```

## Proxy Configuration

If you're using Dify API integration and experiencing timeout issues, you may need to configure a proxy:

```bash
# Set proxy environment variables (replace port with your proxy port)
export HTTP_PROXY=http://127.0.0.1:7897
export HTTPS_PROXY=http://127.0.0.1:7897

# Start service
./event-pod-services api --config ./config.yaml
```

**Alternative**: Enable TUN mode in your proxy software (Clash, Surge, etc.) to automatically proxy all traffic.

For detailed proxy configuration, see [docs/PROXY_CONFIGURATION.md](docs/PROXY_CONFIGURATION.md).

## Security Best Practices

1. **Never commit sensitive data**:
   - `config.yaml` is gitignored
   - Use environment variables for secrets in production

2. **Change default secrets**:
   - JWT secret
   - Database passwords
   - API keys

3. **Protect private keys**:
   - Store blockchain private keys securely
   - Use hardware wallets or key management services in production

4. **CORS configuration**:
   - Set specific allowed origins in production
   - Don't use `*` in production environments

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Resources

- [Go Documentation](https://golang.org/doc/)
- [GORM Documentation](https://gorm.io/docs/)
- [go-ethereum Documentation](https://geth.ethereum.org/docs/)
- [Protocol Buffers](https://protobuf.dev/)
- [Redis Documentation](https://redis.io/docs/)
- [Elasticsearch Documentation](https://www.elastic.co/guide/)

## Support

For questions and support, please open an issue in the GitHub repository.
