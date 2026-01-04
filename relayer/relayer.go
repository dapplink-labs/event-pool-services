package relayer

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"

	"github.com/multimarket-labs/event-pod-services/common/tasks"
	"github.com/multimarket-labs/event-pod-services/database"
	"github.com/multimarket-labs/event-pod-services/metrics"
	"github.com/multimarket-labs/event-pod-services/relayer/driver"
	"github.com/multimarket-labs/event-pod-services/services/websocket"
)

type RelayerProcessor struct {
	LoopInterval     time.Duration
	db               *database.DB
	ethClient        map[uint64]*ethclient.Client
	RrmMangerAddress map[uint64]string
	driverEngine     map[uint64]driver.DriverEngine
	PhoenixMetrics   *metrics.PhoenixMetrics
	resourceCtx      context.Context
	resourceCancel   context.CancelFunc
	tasks            tasks.Group
	wsHub            *websocket.Hub
}

func NewRelayerProcessor(db *database.DB, ethClient map[uint64]*ethclient.Client, rrmMangerAddress map[uint64]string, driverEngine map[uint64]driver.DriverEngine, PhoenixMetrics *metrics.PhoenixMetrics, wsHub *websocket.Hub, shutdown context.CancelCauseFunc) (*RelayerProcessor, error) {
	resCtx, resCancel := context.WithCancel(context.Background())
	return &RelayerProcessor{
		db:               db,
		resourceCtx:      resCtx,
		resourceCancel:   resCancel,
		ethClient:        ethClient,
		RrmMangerAddress: rrmMangerAddress,
		driverEngine:     driverEngine,
		PhoenixMetrics:   PhoenixMetrics,
		wsHub:            wsHub,
		tasks: tasks.Group{HandleCrit: func(err error) {
			shutdown(fmt.Errorf("critical error in worker handle processor: %w", err))
		}},
	}, nil
}

func (sh *RelayerProcessor) Close() error {
	sh.resourceCancel()
	return sh.tasks.Wait()
}

func (sh *RelayerProcessor) Start() error {
	workerTicker := time.NewTicker(time.Second * 5)
	sh.tasks.Go(func() error {
		for range workerTicker.C {
			log.Info("==========start handle==========")
		}
		return nil
	})

	sh.tasks.Go(func() error {
		for range workerTicker.C {
			log.Info("==========start handle==========")
		}
		return nil
	})
	return nil
}
