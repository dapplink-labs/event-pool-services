package relayer_node

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"

	"github.com/multimarket-labs/event-pod-services/common/httputil"
	"github.com/multimarket-labs/event-pod-services/config"
	"github.com/multimarket-labs/event-pod-services/database"
	"github.com/multimarket-labs/event-pod-services/metrics"
	"github.com/multimarket-labs/event-pod-services/relayer"
	"github.com/multimarket-labs/event-pod-services/relayer/driver"
	"github.com/multimarket-labs/event-pod-services/services/websocket"
	"github.com/multimarket-labs/event-pod-services/worker"
)

type PhoenixNode struct {
	DB               *database.DB
	metricsServer    *httputil.HTTPServer
	metricsRegistry  *prometheus.Registry
	phoenixMetrics   *metrics.PhoenixMetrics
	relayerProcessor *relayer.RelayerProcessor
	WorkerHandle     *worker.WorkerHandle
	wsHub            *websocket.Hub
	wsServer         *httputil.HTTPServer
	shutdown         context.CancelCauseFunc
	stopped          atomic.Bool
	chainIdList      []uint64
}

type RpcServerConfig struct {
	GrpcHostname string
	GrpcPort     int
}

func NewPhoenixNode(ctx context.Context, cfg *config.Config, shutdown context.CancelCauseFunc) (*PhoenixNode, error) {
	log.Info("New phoenix services start️ 🕖")

	metricsRegistry := metrics.NewRegistry()

	PhoenixMetrics := metrics.NewPhoenixMetrics(metricsRegistry, "phoenix")

	out := &PhoenixNode{
		metricsRegistry: metricsRegistry,
		phoenixMetrics:  PhoenixMetrics,
		shutdown:        shutdown,
	}
	if err := out.initFromConfig(ctx, cfg); err != nil {
		return nil, errors.Join(err, out.Stop(ctx))
	}
	log.Info("New phoenix services success🏅️")
	return out, nil
}

func (as *PhoenixNode) Start(ctx context.Context) error {
	errPhoenix := as.relayerProcessor.Start()
	if errPhoenix != nil {
		log.Error("start relayer processor fail", "err", errPhoenix)
		return errPhoenix
	}

	errWorker := as.WorkerHandle.Start()
	if errWorker != nil {
		log.Error("start worker handle fail", "err", errWorker)
		return errWorker
	}
	return nil
}

func (as *PhoenixNode) Stop(ctx context.Context) error {
	var result error
	if as.DB != nil {
		if err := as.DB.Close(); err != nil {
			result = errors.Join(result, fmt.Errorf("failed to close DB: %w", err))
		}
	}

	if as.metricsServer != nil {
		if err := as.metricsServer.Close(); err != nil {
			result = errors.Join(result, fmt.Errorf("failed to close metrics server: %w", err))
		}
	}

	if as.wsHub != nil {
		as.wsHub.CloseAllClients()
	}

	if as.wsServer != nil {
		if err := as.wsServer.Stop(ctx); err != nil {
			result = errors.Join(result, fmt.Errorf("failed to stop WebSocket server: %w", err))
		}
	}

	as.stopped.Store(true)

	log.Info("phoenix services stopped")

	return result
}

func (as *PhoenixNode) Stopped() bool {
	return as.stopped.Load()
}

func (as *PhoenixNode) initFromConfig(ctx context.Context, cfg *config.Config) error {
	if err := as.initDB(ctx, cfg.MasterDB); err != nil {
		return fmt.Errorf("failed to init DB: %w", err)
	}

	as.wsHub = websocket.NewHub()
	go as.wsHub.Run()

	if err := as.startWebSocketServer(cfg.WebsocketServer); err != nil {
		return fmt.Errorf("failed to start web socket server: %w", err)
	}

	if err := as.initRelayer(cfg); err != nil {
		return fmt.Errorf("failed to init relayer processor: %w", err)
	}

	if err := as.initWorker(cfg); err != nil {
		return fmt.Errorf("failed to init worker processor: %w", err)
	}

	err := as.startMetricsServer(cfg.MetricsServer)
	if err != nil {
		log.Error("start metrics server fail", "err", err)
		return err
	}
	return nil
}

func (as *PhoenixNode) startWebSocketServer(serverConfig config.ServerConfig) error {
	addr := net.JoinHostPort(serverConfig.Host, strconv.Itoa(serverConfig.Port))

	wsRouter := chi.NewRouter()
	wsRouter.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWebSocket(as.wsHub, w, r)
	})

	srv, err := httputil.StartHTTPServer(addr, wsRouter)
	if err != nil {
		return fmt.Errorf("failed to start WebSocket server: %w", err)
	}
	log.Info("WebSocket server started", "addr", srv.Addr().String())
	as.wsServer = srv
	return nil
}

func (as *PhoenixNode) initDB(ctx context.Context, cfg config.DBConfig) error {
	db, err := database.NewDB(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	as.DB = db
	log.Info("Init database success")
	return nil
}

func (as *PhoenixNode) initWorker(config *config.Config) error {
	var chainIds []string
	for i := range config.RPCs {
		chainIds = append(chainIds, strconv.Itoa(int(config.RPCs[i].ChainId)))
	}
	wkConfig := &worker.WorkerHandleConfig{
		LoopInterval: time.Second * 5,
		ChainIds:     chainIds,
	}
	workerHandle, err := worker.NewWorkerHandle(as.DB, wkConfig, as.wsHub, as.shutdown)
	if err != nil {
		return err
	}
	as.WorkerHandle = workerHandle
	return nil
}

func (as *PhoenixNode) startMetricsServer(cfg config.ServerConfig) error {
	srv, err := metrics.StartServer(as.metricsRegistry, cfg.Host, cfg.Port)
	if err != nil {
		return fmt.Errorf("metrics server failed to start: %w", err)
	}
	as.metricsServer = srv
	log.Info("metrics server started", "port", cfg.Port, "addr", srv.Addr())
	return nil
}

func (as *PhoenixNode) initRelayer(config *config.Config) error {
	var ethClient map[uint64]*ethclient.Client
	var poolMangerAddress map[uint64]string
	var driverEngine map[uint64]driver.DriverEngine

	for i := range config.RPCs {

		rpcItem := config.RPCs[i]

		ethClt, err := driver.EthClientWithTimeout(context.Background(), rpcItem.RpcUrl)
		if err != nil {
			log.Error("new eth client fail", "err", err)
			return err
		}

		ecdsaPrivateKey, err := crypto.HexToECDSA(config.PrivateKey)
		if err != nil {
			log.Error("ecdsa format fail", "err", err)
			return err
		}
		log.Info("init relayer start", "chainId", config.RPCs[i].ChainId, "RpcUrl", rpcItem.RpcUrl)

		if ethClient == nil {
			ethClient = make(map[uint64]*ethclient.Client)
		}
		ethClient[rpcItem.ChainId] = ethClt

		if poolMangerAddress == nil {
			poolMangerAddress = make(map[uint64]string)
		}
		poolMangerAddress[rpcItem.ChainId] = rpcItem.Contracts.ReferralRewardManager

		dregConf := &driver.DriverEngineConfig{
			ChainClient:               ethClt,
			ChainId:                   big.NewInt(int64(rpcItem.ChainId)),
			ContractAddress:           common.HexToAddress(rpcItem.Contracts.ReferralRewardManager),
			CallerAddress:             common.HexToAddress(config.CallerAddress),
			PrivateKey:                ecdsaPrivateKey,
			NumConfirmations:          config.NumConfirmations,
			SafeAbortNonceTooLowCount: config.SafeAbortNonceTooLowCount,
		}

		drEngine, err := driver.NewDriverEngine(context.Background(), as.phoenixMetrics, dregConf)
		if err != nil {
			log.Error("new drive engine fail", "err", err)
			return err
		}
		if driverEngine == nil {
			driverEngine = make(map[uint64]driver.DriverEngine)
		}
		driverEngine[rpcItem.ChainId] = *drEngine
	}
	relayerProcessor, err := relayer.NewRelayerProcessor(as.DB, ethClient, poolMangerAddress, driverEngine, as.phoenixMetrics, as.wsHub, as.shutdown)
	if err != nil {
		log.Error("new relayer processor fail", "err", err)
		return err
	}
	as.relayerProcessor = relayerProcessor
	return nil
}
