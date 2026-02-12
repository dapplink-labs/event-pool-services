package crawler

import (
	"context"
	"fmt"
	"github.com/multimarket-labs/event-pod-services/common/tasks"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/multimarket-labs/event-pod-services/crawler/crypto"
	"github.com/multimarket-labs/event-pod-services/crawler/sports"
	"github.com/multimarket-labs/event-pod-services/database"
)

type CrawlerConfig struct {
	LoopInterval  time.Duration
	NBAConfig     sports.NBACrawlerConfig
	BinanceConfig crypto.BinanceCrawlerConfig
	BybitConfig   crypto.BybitCrawlerConfig
}

type Crawler struct {
	db             *database.DB
	wConf          *CrawlerConfig
	nbaCrawler     *sports.NBACrawler
	binanceCrawler *crypto.BinanceCrawler
	bybitCrawler   *crypto.BybitCrawler
	stopCh         chan struct{}
	tasks          tasks.Group
}

func NewCrawler(db *database.DB, wConf *CrawlerConfig, shutdown context.CancelCauseFunc) (*Crawler, error) {
	nbaCrawler, err := sports.NewNBACrawler(db, wConf.NBAConfig)
	if err != nil {
		log.Warn("Failed to initialize NBA crawler, NBA sync will be skipped", "err", err)
	}

	binanceCrawler, err := crypto.NewBinanceCrawler(db, wConf.BinanceConfig)
	if err != nil {
		log.Warn("Failed to initialize Binance crawler, Binance sync will be skipped", "err", err)
	}

	bybitCrawler, err := crypto.NewBybitCrawler(db, wConf.BybitConfig)
	if err != nil {
		log.Warn("Failed to initialize Bybit crawler, Bybit sync will be skipped", "err", err)
	}

	return &Crawler{
		db:             db,
		wConf:          wConf,
		nbaCrawler:     nbaCrawler,
		binanceCrawler: binanceCrawler,
		bybitCrawler:   bybitCrawler,
		stopCh:         make(chan struct{}),
		tasks: tasks.Group{
			HandleCrit: func(err error) {
				shutdown(fmt.Errorf("critical error in worker handle processor: %w", err))
			},
		},
	}, nil
}

func (sh *Crawler) Close() error {
	close(sh.stopCh)
	if sh.binanceCrawler != nil {
		sh.binanceCrawler.StopWebSocketStream()
	}
	if sh.bybitCrawler != nil {
		sh.bybitCrawler.StopWebSocketStream()
	}
	return nil
}

func (sh *Crawler) Start() error {

	go sh.runNBASync()
	go sh.runBinanceSync()
	go sh.runBybitSync()

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

func (sh *Crawler) runBinanceSync() {
	ctx := context.Background()

	if sh.binanceCrawler != nil {
		if err := sh.binanceCrawler.StartWebSocketStream(ctx); err != nil {
			log.Error("Failed to start Binance WebSocket stream, falling back to REST API", "err", err)
			sh.syncBinancePrices(ctx)
			workerTicker := time.NewTicker(30 * time.Second)
			sh.tasks.Go(func() error {
				for range workerTicker.C {
					sh.syncBinancePrices(ctx)
				}
				return nil
			})
		} else {
			log.Info("Binance WebSocket stream started for real-time price updates")
		}
	}
}

func (sh *Crawler) syncBinancePrices(ctx context.Context) {
	if sh.binanceCrawler == nil {
		return
	}

	if err := sh.binanceCrawler.SyncPrices(ctx); err != nil {
		log.Error("Failed to sync Binance prices", "err", err)
	}
}

func (sh *Crawler) runBybitSync() {
	ctx := context.Background()

	if sh.bybitCrawler != nil {
		if err := sh.bybitCrawler.StartWebSocketStream(ctx); err != nil {
			log.Error("Failed to start Bybit WebSocket stream", "err", err)
		} else {
			log.Info("Bybit WebSocket stream started for real-time price updates")
		}
	}
}
