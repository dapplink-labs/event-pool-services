package worker

import (
	"context"
	"fmt"

	"time"

	"github.com/ethereum/go-ethereum/log"

	"github.com/multimarket-labs/event-pod-services/common/tasks"
	"github.com/multimarket-labs/event-pod-services/database"
	"github.com/multimarket-labs/event-pod-services/services/websocket"
)

type WorkerHandleConfig struct {
	LoopInterval time.Duration
	ChainIds     []string
}

type WorkerHandle struct {
	db             *database.DB
	wConf          *WorkerHandleConfig
	wsHub          *websocket.Hub
	resourceCtx    context.Context
	resourceCancel context.CancelFunc
	tasks          tasks.Group
}

func NewWorkerHandle(db *database.DB, wConf *WorkerHandleConfig, wsHub *websocket.Hub, shutdown context.CancelCauseFunc) (*WorkerHandle, error) {
	resCtx, resCancel := context.WithCancel(context.Background())
	return &WorkerHandle{
		db:             db,
		wConf:          wConf,
		wsHub:          wsHub,
		resourceCtx:    resCtx,
		resourceCancel: resCancel,
		tasks: tasks.Group{
			HandleCrit: func(err error) {
				shutdown(fmt.Errorf("critical error in worker handle processor: %w", err))
			},
		},
	}, nil
}

func (sh *WorkerHandle) Close() error {
	sh.resourceCancel()
	return sh.tasks.Wait()
}

func (sh *WorkerHandle) Start() error {
	workerTicker := time.NewTicker(sh.wConf.LoopInterval)
	sh.tasks.Go(func() error {
		for range workerTicker.C {
			log.Info("==== star ======")
		}
		return nil
	})

	sh.tasks.Go(func() error {
		for range workerTicker.C {
			log.Info("==== star ======")
		}
		return nil
	})
	return nil
}
