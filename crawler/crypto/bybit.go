package crypto

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"

	bybit "github.com/bybit-exchange/bybit.go.api"

	"github.com/multimarket-labs/event-pod-services/database"
)

type BybitCrawler struct {
	db              *database.DB
	Category        string
	Ecosystem       string
	Period          string
	defaultLanguage string
	symbols         []string
	wsClient        *bybit.WebSocket
	wsMutex         sync.RWMutex
	wsStopCh        chan struct{}
	wsRunning       bool
	wsUrl           string
}

type BybitTickerPrice struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

type BybitTickerStream struct {
	Topic string `json:"topic"`
	Type  string `json:"type"`
	Ts    int64  `json:"ts"`
	Data  struct {
		Symbol       string `json:"symbol"`
		LastPrice    string `json:"lastPrice"`
		OpenPrice    string `json:"openPrice"`
		HighPrice    string `json:"highPrice"`
		LowPrice     string `json:"lowPrice"`
		PrevPrice24h string `json:"prevPrice24h"`
		Volume24h    string `json:"volume24h"`
		Turnover24h  string `json:"turnover24h"`
		Price24hPcnt string `json:"price24hPcnt"`
		Change24h    string `json:"change24h"`
	} `json:"data"`
}

type BybitCrawlerConfig struct {
	CategoryGUID  string
	EcosystemGUID string
	PeriodGUID    string
	LanguageGUID  string
	Symbols       []string
	WsUrl         string
}

func NewBybitCrawler(db *database.DB, config BybitCrawlerConfig) (*BybitCrawler, error) {
	return &BybitCrawler{
		db:              db,
		Category:        config.CategoryGUID,
		Ecosystem:       config.EcosystemGUID,
		Period:          config.PeriodGUID,
		defaultLanguage: config.LanguageGUID,
		symbols:         config.Symbols,
		wsStopCh:        make(chan struct{}),
		wsRunning:       false,
		wsUrl:           config.WsUrl,
	}, nil
}

func (c *BybitCrawler) StartWebSocketStream(ctx context.Context) error {
	c.wsMutex.Lock()
	if c.wsRunning {
		c.wsMutex.Unlock()
		return fmt.Errorf("WebSocket stream is already running")
	}
	c.wsRunning = true
	c.wsMutex.Unlock()

	go c.runWebSocketStream(ctx)
	return nil
}

func (c *BybitCrawler) StopWebSocketStream() {
	c.wsMutex.Lock()
	defer c.wsMutex.Unlock()

	if !c.wsRunning {
		return
	}

	close(c.wsStopCh)
	if c.wsClient != nil {
		c.wsClient.Disconnect()
		c.wsClient = nil
	}
	c.wsRunning = false
	log.Info("Bybit WebSocket stream stopped")
}

func (c *BybitCrawler) runWebSocketStream(ctx context.Context) {
	reconnectDelay := 5 * time.Second
	maxReconnectDelay := 60 * time.Second

	for {
		select {
		case <-c.wsStopCh:
			return
		case <-ctx.Done():
			return
		default:
			if err := c.connectAndSubscribe(ctx); err != nil {
				log.Error("WebSocket connection error, will reconnect", "err", err, "delay", reconnectDelay)
				time.Sleep(reconnectDelay)
				if reconnectDelay < maxReconnectDelay {
					reconnectDelay = time.Duration(float64(reconnectDelay) * 1.5)
				}
				continue
			}
			reconnectDelay = 5 * time.Second
		}
	}
}

func (c *BybitCrawler) connectAndSubscribe(ctx context.Context) error {

	if len(c.symbols) == 0 {
		return fmt.Errorf("no symbols configured for Bybit crawler")
	}

	topics := make([]string, len(c.symbols))
	for i, symbol := range c.symbols {
		topics[i] = fmt.Sprintf("tickers.%s", strings.ToUpper(symbol))
	}

	log.Info("Attempting Bybit WebSocket connection", "endpoint", c.wsUrl, "topics", topics)

	wsClient := bybit.NewBybitPublicWebSocket(c.wsUrl, func(message string) error {
		return c.handleWebSocketMessage([]byte(message))
	})

	connectedWS := wsClient.Connect()
	if connectedWS == nil {
		return fmt.Errorf("failed to connect to Bybit WebSocket: returned nil")
	}

	_, err := connectedWS.SendSubscription(topics)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topics: %w", err)
	}

	c.wsMutex.Lock()
	c.wsClient = connectedWS
	c.wsMutex.Unlock()

	log.Info("Bybit WebSocket connected and subscribed", "topics", topics)

	select {
	case <-ctx.Done():
		if c.wsClient != nil {
			c.wsClient.Disconnect()
		}
		return nil
	case <-c.wsStopCh:
		if c.wsClient != nil {
			c.wsClient.Disconnect()
		}
		return nil
	}
}

func (c *BybitCrawler) handleWebSocketMessage(message []byte) error {
	var genericMsg map[string]interface{}
	if err := json.Unmarshal(message, &genericMsg); err != nil {
		log.Debug("Failed to parse message as JSON", "err", err, "message", string(message))
		return nil
	}

	topic, ok := genericMsg["topic"].(string)
	if !ok || !strings.HasPrefix(topic, "tickers.") {
		log.Debug("Received non-ticker message", "topic", topic, "message", string(message))
		return nil
	}

	var tickerStream BybitTickerStream
	if err := json.Unmarshal(message, &tickerStream); err != nil {
		log.Warn("Failed to unmarshal ticker stream", "err", err, "message", string(message))
		return nil
	}

	lastPrice := tickerStream.Data.LastPrice
	if lastPrice == "" {
		if data, ok := genericMsg["data"].(map[string]interface{}); ok {
			if price, ok := data["lastPrice"].(string); ok && price != "" {
				lastPrice = price
			} else if price, ok := data["last_price"].(string); ok && price != "" {
				lastPrice = price
			}
		}
		if lastPrice == "" {
			log.Debug("Empty price in ticker data, skipping", "message", string(message))
			return nil
		}
	}

	symbol := tickerStream.Data.Symbol
	if symbol == "" {
		parts := strings.Split(topic, ".")
		if len(parts) >= 2 {
			symbol = parts[1]
		}
		if symbol == "" {
			log.Warn("Cannot extract symbol from message", "topic", topic, "message", string(message))
			return nil
		}
	}

	priceData := &BybitTickerPrice{
		Symbol: symbol,
		Price:  lastPrice,
	}

	if err := c.processPrice(priceData); err != nil {
		log.Error("Failed to process price update", "err", err, "symbol", symbol, "price", lastPrice)
		return nil
	}

	log.Debug("Processed real-time price update", "symbol", priceData.Symbol, "price", priceData.Price)
	return nil
}

func (c *BybitCrawler) processPrice(priceData *BybitTickerPrice) error {
	return c.db.Transaction(func(txDB *database.DB) error {
		externalID := fmt.Sprintf("BYBIT_%s", priceData.Symbol)

		existingEvent, err := txDB.Event.GetEventByExternalID(externalID)
		if err != nil {
			return fmt.Errorf("failed to check existing event: %w", err)
		}

		price := priceData.Price
		mainScore := "0"
		clusterScore := "0"

		now := time.Now()
		dateCode := now.Format("2006-01-02")
		periodCode := fmt.Sprintf("CRYPTO_BYBIY_%s", dateCode)

		eventPeriodGUID, err := c.ensureEventPeriod(txDB, periodCode, dateCode)
		if err != nil {
			return fmt.Errorf("failed to ensure event period: %w", err)
		}

		eventInfo := database.JSONB{
			"external_id": externalID,
			"symbol":      priceData.Symbol,
			"price":       priceData.Price,
			"timestamp":   time.Now().Unix(),
		}

		eventTitle := fmt.Sprintf("%s Price", priceData.Symbol)

		if existingEvent != nil {
			updates := map[string]interface{}{
				"price":             price,
				"main_score":        mainScore,
				"cluster_score":     clusterScore,
				"is_live":           int16(0),
				"stage":             "LIVE",
				"info":              eventInfo,
				"event_period_guid": eventPeriodGUID,
				"updated_at":        time.Now(),
			}

			if err := txDB.Event.UpdateEventFields(existingEvent.GUID, updates); err != nil {
				return fmt.Errorf("failed to update event: %w", err)
			}

			if err := c.updateEventLanguage(txDB, existingEvent.GUID, eventTitle); err != nil {
				log.Warn("Failed to update event language", "err", err)
			}
		} else {
			event := &database.Event{
				CategoryGUID:         c.Category,
				EcosystemGUID:        c.Ecosystem,
				EventPeriodGUID:      eventPeriodGUID,
				MainTeamGroupGUID:    "0",
				ClusterTeamGroupGUID: "0",
				ExternalId:           externalID,
				MainScore:            mainScore,
				ClusterScore:         clusterScore,
				Price:                price,
				Logo:                 "",
				EventType:            0,
				ExperimentResult:     "",
				Info:                 eventInfo,
				IsOnline:             true,
				IsLive:               int16(0),
				IsSports:             false,
				Stage:                "LIVE",
			}

			if err := txDB.Event.CreateEvent(event); err != nil {
				return fmt.Errorf("failed to create event: %w", err)
			}

			createdEvent, err := txDB.Event.GetEventByExternalID(externalID)
			if err != nil {
				return fmt.Errorf("failed to retrieve created event: %w", err)
			}
			if createdEvent == nil {
				return fmt.Errorf("created event not found by external_id: %s", externalID)
			}

			eventLang := &database.EventLanguage{
				EventGUID:    createdEvent.GUID,
				LanguageGUID: c.defaultLanguage,
				Title:        eventTitle,
				Rules:        fmt.Sprintf("Real-time price tracking for %s on Bybit", priceData.Symbol),
			}

			if err := txDB.Event.CreateEventLanguage(eventLang); err != nil {
				return fmt.Errorf("failed to create event language: %w", err)
			}

			log.Info("Created Bybit event", "symbol", priceData.Symbol, "event_guid", createdEvent.GUID)
		}

		return nil
	})
}

func (c *BybitCrawler) ensureEventPeriod(txDB *database.DB, periodCode, dateCode string) (string, error) {
	existingPeriod, err := txDB.EventPeriod.GetEventPeriodByCode(periodCode)
	if err != nil {
		return "", fmt.Errorf("failed to check existing event period: %w", err)
	}

	if existingPeriod != nil {
		return existingPeriod.GUID, nil
	}

	now := time.Now()
	scheduled := now.Format("2006-01-02 15:04:05")

	newPeriod := &database.EventPeriod{
		Code:      periodCode,
		IsActive:  true,
		Scheduled: scheduled,
		Remark:    fmt.Sprintf("Crypto price date: %s", dateCode),
		Extra:     database.JSONB{},
	}

	if err := txDB.EventPeriod.CreateEventPeriod(newPeriod); err != nil {
		return "", fmt.Errorf("failed to create event period: %w", err)
	}

	createdPeriod, err := txDB.EventPeriod.GetEventPeriodByCode(periodCode)
	if err != nil || createdPeriod == nil {
		return "", fmt.Errorf("failed to retrieve created event period: %w", err)
	}

	log.Info("Created event period", "code", periodCode, "guid", createdPeriod.GUID)
	return createdPeriod.GUID, nil
}

func (c *BybitCrawler) updateEventLanguage(txDB *database.DB, eventGUID, title string) error {
	eventLang, err := txDB.Event.GetEventLanguage(eventGUID, c.defaultLanguage)
	if err != nil {
		return err
	}

	if eventLang == nil {
		newEventLang := &database.EventLanguage{
			EventGUID:    eventGUID,
			LanguageGUID: c.defaultLanguage,
			Title:        title,
			Rules:        "",
		}
		return txDB.Event.CreateEventLanguage(newEventLang)
	}

	eventLang.Title = title
	return txDB.Event.UpdateEventLanguage(eventLang)
}
