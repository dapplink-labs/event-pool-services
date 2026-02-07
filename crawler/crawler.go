package crawler

import (
	"time"

	"github.com/multimarket-labs/event-pod-services/database"
)

type CrawlerConfig struct {
	LoopInterval time.Duration
	ChainIds     []string
}

type Crawler struct {
	db    *database.DB
	wConf *CrawlerConfig
}

func NewCrawler(db *database.DB, wConf *CrawlerConfig) (*Crawler, error) {
	return &Crawler{
		db:    db,
		wConf: wConf,
	}, nil
}

func (sh *Crawler) Close() error {
	return nil
}

func (sh *Crawler) Start() error {
	return nil
}
