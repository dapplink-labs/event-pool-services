package crawler

import (
	"context"
	"fmt"
	"github.com/multimarket-labs/event-pod-services/common/tasks"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/multimarket-labs/event-pod-services/crawler/sports"
	"github.com/multimarket-labs/event-pod-services/database"
)

type CrawlerConfig struct {
	LoopInterval time.Duration
	NBAConfig    sports.NBACrawlerConfig
}

type Crawler struct {
	db         *database.DB
	wConf      *CrawlerConfig
	nbaCrawler *sports.NBACrawler
	stopCh     chan struct{}
	tasks      tasks.Group
}

func NewCrawler(db *database.DB, wConf *CrawlerConfig, shutdown context.CancelCauseFunc) (*Crawler, error) {
	nbaCrawler, err := sports.NewNBACrawler(db, wConf.NBAConfig)
	if err != nil {
		log.Warn("Failed to initialize NBA crawler, NBA sync will be skipped", "err", err)
	}

	return &Crawler{
		db:         db,
		wConf:      wConf,
		nbaCrawler: nbaCrawler,
		stopCh:     make(chan struct{}),
		tasks: tasks.Group{
			HandleCrit: func(err error) {
				shutdown(fmt.Errorf("critical error in worker handle processor: %w", err))
			},
		},
	}, nil
}

func (sh *Crawler) Close() error {
	close(sh.stopCh)
	return nil
}

func (sh *Crawler) Start() error {

	go sh.runNBASync()

	log.Info("Crawler service started")
	return nil
}

func (sh *Crawler) runNBASync() {
	ctx := context.Background()
	sh.syncNBAToday(ctx)

	workerTicker := time.NewTicker(5 * time.Minute)
	sh.tasks.Go(func() error {
		for range workerTicker.C {
			sh.syncNBAToday(ctx)
		}
		return nil
	})

}

func (sh *Crawler) syncNBAToday(ctx context.Context) {
	if sh.nbaCrawler == nil {
		return
	}

	if err := sh.nbaCrawler.SyncToday(ctx); err != nil {
		log.Error("Failed to sync today's NBA schedule", "err", err)
	}
}

func (sh *Crawler) syncNBADateRange(ctx context.Context) {
	if sh.nbaCrawler == nil {
		return
	}

	now := time.Now()
	startDate := now
	endDate := now.AddDate(0, 0, 7)

	if err := sh.nbaCrawler.SyncDateRange(ctx, startDate, endDate); err != nil {
		log.Error("Failed to sync NBA schedule range", "start", startDate, "end", endDate, "err", err)
	}
}
