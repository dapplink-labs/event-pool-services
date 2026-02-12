package relayer_node

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/ethereum/go-ethereum/log"

	"github.com/multimarket-labs/event-pod-services/common/httputil"
	"github.com/multimarket-labs/event-pod-services/config"
	"github.com/multimarket-labs/event-pod-services/crawler"
	"github.com/multimarket-labs/event-pod-services/crawler/crypto"
	"github.com/multimarket-labs/event-pod-services/crawler/sports"
	"github.com/multimarket-labs/event-pod-services/database"
	"github.com/multimarket-labs/event-pod-services/metrics"
)

type EventPool struct {
	DB               *database.DB
	metricsServer    *httputil.HTTPServer
	metricsRegistry  *prometheus.Registry
	eventPoolMetrics *metrics.EventPoolMetrics
	Crawler          *crawler.Crawler
	wsServer         *httputil.HTTPServer
	shutdown         context.CancelCauseFunc
	stopped          atomic.Bool
}

type RpcServerConfig struct {
	GrpcHostname string
	GrpcPort     int
}

func NewEventPool(ctx context.Context, cfg *config.Config, shutdown context.CancelCauseFunc) (*EventPool, error) {
	log.Info("New event pool services start️ 🕖")

	metricsRegistry := metrics.NewRegistry()

	EventPoolMetrics := metrics.NewEventPoolMetrics(metricsRegistry, "eventPool")

	out := &EventPool{
		metricsRegistry:  metricsRegistry,
		eventPoolMetrics: EventPoolMetrics,
		shutdown:         shutdown,
	}
	if err := out.initFromConfig(ctx, cfg); err != nil {
		return nil, errors.Join(err, out.Stop(ctx))
	}
	log.Info("new event pool services success🏅️")
	return out, nil
}

func (as *EventPool) Start(ctx context.Context) error {
	errWorker := as.Crawler.Start()
	if errWorker != nil {
		log.Error("start crawler handle fail", "err", errWorker)
		return errWorker
	}
	return nil
}

func (as *EventPool) Stop(ctx context.Context) error {
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

	if as.wsServer != nil {
		if err := as.wsServer.Stop(ctx); err != nil {
			result = errors.Join(result, fmt.Errorf("failed to stop WebSocket server: %w", err))
		}
	}

	as.stopped.Store(true)

	log.Info("event pool services stopped")

	return result
}

func (as *EventPool) Stopped() bool {
	return as.stopped.Load()
}

func (as *EventPool) initFromConfig(ctx context.Context, cfg *config.Config) error {
	if err := as.initDB(ctx, cfg.MasterDB); err != nil {
		return fmt.Errorf("failed to init DB: %w", err)
	}

	if err := as.startWebSocketServer(cfg.WebsocketServer); err != nil {
		return fmt.Errorf("failed to start web socket server: %w", err)
	}

	if err := as.initWorker(cfg, as.shutdown); err != nil {
		return fmt.Errorf("failed to init crawler processor: %w", err)
	}

	err := as.startMetricsServer(cfg.MetricsServer)
	if err != nil {
		log.Error("start metrics server fail", "err", err)
		return err
	}
	return nil
}

func (as *EventPool) startWebSocketServer(serverConfig config.ServerConfig) error {
	addr := net.JoinHostPort(serverConfig.Host, strconv.Itoa(serverConfig.Port))

	wsRouter := chi.NewRouter()

	srv, err := httputil.StartHTTPServer(addr, wsRouter)
	if err != nil {
		return fmt.Errorf("failed to start WebSocket server: %w", err)
	}
	log.Info("WebSocket server started", "addr", srv.Addr().String())
	as.wsServer = srv
	return nil
}

func (as *EventPool) initDB(ctx context.Context, cfg config.DBConfig) error {
	db, err := database.NewDB(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	as.DB = db
	log.Info("Init database success")
	return nil
}

func (as *EventPool) initWorker(config *config.Config, shutdown context.CancelCauseFunc) error {
	allLanguages, err := as.DB.Languages.QueryAllLanguages()
	if err != nil {
		log.Warn("Failed to get language list, using default language", "err", err)
		allLanguages = []database.Languages{}
	}

	var languages []string
	var defaultLangGUID string
	for _, lang := range allLanguages {
		languages = append(languages, lang.LanguageName)
		if lang.IsDefault || lang.LanguageName == "en" {
			defaultLangGUID = lang.GUID
		}
	}
	if len(languages) == 0 {
		languages = []string{"en", "zh"}
	}

	// Dynamically get category and ecosystem GUIDs from database
	sportCategory, err := as.DB.Category.GetCategoryByCode("SPORT")
	if err != nil {
		log.Warn("Failed to get sport category, NBA crawler may not work properly", "err", err)
	}

	cryptoCategory, err := as.DB.Category.GetCategoryByCode("CRYPTO")
	if err != nil {
		log.Warn("Failed to get crypto category, Binance crawler may not work properly", "err", err)
	}

	nbaEcosystem, err := as.DB.Ecosystem.GetEcosystemByCode("NBA")
	if err != nil {
		log.Warn("Failed to get NBA ecosystem, NBA crawler may not work properly", "err", err)
	}

	binanceEcosystem, err := as.DB.Ecosystem.GetEcosystemByCode("BINANCE")
	if err != nil {
		log.Warn("Failed to get Binance ecosystem, Binance crawler may not work properly", "err", err)
	}

	bybitEcosystem, err := as.DB.Ecosystem.GetEcosystemByCode("BYBIT")
	if err != nil {
		log.Warn("Failed to get Bybit ecosystem, Bybit crawler may not work properly", "err", err)
	}

	okxEcosystem, err := as.DB.Ecosystem.GetEcosystemByCode("OKX")
	if err != nil {
		log.Warn("Failed to get OKX ecosystem, OKX crawler may not work properly", "err", err)
	}

	nbaConfig := sports.NBACrawlerConfig{
		AccessLevel:   config.Sportradar.AccessLevel,
		ApiKey:        config.Sportradar.ApiKey,
		LanguageCode:  languages,
		CategoryGUID:  sportCategory.GUID,
		EcosystemGUID: nbaEcosystem.GUID,
		PeriodGUID:    "",
		LanguageGUID:  defaultLangGUID,
	}

	binanceConfig := crypto.BinanceCrawlerConfig{
		CategoryGUID:  cryptoCategory.GUID,
		EcosystemGUID: binanceEcosystem.GUID,
		PeriodGUID:    "",
		LanguageGUID:  defaultLangGUID,
		Symbols:       config.Binance.Symbols,
		ApiUrl:        config.Binance.ApiURL,
		WsUrl:         config.Binance.WsUrl,
	}

	bybitConfig := crypto.BybitCrawlerConfig{
		CategoryGUID:  cryptoCategory.GUID,
		EcosystemGUID: bybitEcosystem.GUID,
		PeriodGUID:    "",
		LanguageGUID:  defaultLangGUID,
		Symbols:       config.Bybit.Symbols,
		WsUrl:         config.Bybit.WsUrl,
	}

	okxConfig := crypto.OKXCrawlerConfig{
		CategoryGUID:  cryptoCategory.GUID,
		EcosystemGUID: okxEcosystem.GUID,
		PeriodGUID:    "",
		LanguageGUID:  defaultLangGUID,
		Symbols:       config.OKX.Symbols,
		WsUrl:         config.OKX.WsUrl,
	}

	wkConfig := &crawler.CrawlerConfig{
		LoopInterval:  time.Second * 5,
		NBAConfig:     nbaConfig,
		BinanceConfig: binanceConfig,
		BybitConfig:   bybitConfig,
		OKXConfig:     okxConfig,
	}
	workerHandle, err := crawler.NewCrawler(as.DB, wkConfig, shutdown)
	if err != nil {
		return err
	}
	as.Crawler = workerHandle
	return nil
}

func (as *EventPool) startMetricsServer(cfg config.ServerConfig) error {
	srv, err := metrics.StartServer(as.metricsRegistry, cfg.Host, cfg.Port)
	if err != nil {
		return fmt.Errorf("metrics server failed to start: %w", err)
	}
	as.metricsServer = srv
	log.Info("metrics server started", "port", cfg.Port, "addr", srv.Addr())
	return nil
}
