package crypto

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/gorilla/websocket"

	"github.com/multimarket-labs/event-pod-services/common/retry"
	"github.com/multimarket-labs/event-pod-services/database"
)

type BinanceCrawler struct {
	db              *database.DB
	baseURL         string
	wsURL           string
	httpClient      *http.Client
	Category        string
	Ecosystem       string
	Period          string
	defaultLanguage string
	symbols         []string
	wsConn          *websocket.Conn
	wsMutex         sync.RWMutex
	wsStopCh        chan struct{}
	wsRunning       bool
}

type BinanceTickerPrice struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

type BinanceTickerStream struct {
	Stream string `json:"stream"`
	Data   struct {
		EventType           string `json:"e"`
		EventTime           int64  `json:"E"`
		Symbol              string `json:"s"`
		PriceChange         string `json:"p"`
		PriceChangePercent  string `json:"P"`
		WeightedAvgPrice    string `json:"w"`
		LastPrice           string `json:"c"`
		LastQuantity        string `json:"Q"`
		OpenPrice           string `json:"o"`
		HighPrice           string `json:"h"`
		LowPrice            string `json:"l"`
		TotalTradedVolume   string `json:"v"`
		TotalTradedQuoteVol string `json:"q"`
		OpenTime            int64  `json:"O"`
		CloseTime           int64  `json:"C"`
		FirstTradeID        int64  `json:"F"`
		LastTradeID         int64  `json:"L"`
		TotalNumberOfTrades int64  `json:"n"`
	} `json:"data"`
}

type BinanceWSMessage struct {
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
	ID     interface{}   `json:"id"`
	Result interface{}   `json:"result,omitempty"`
}

type BinanceCrawlerConfig struct {
	CategoryGUID  string
	EcosystemGUID string
	PeriodGUID    string
	LanguageGUID  string
	Symbols       []string // Trading pairs to monitor, e.g., ["BTCUSDT", "ETHUSDT", "BNBUSDT"]
	ApiUrl        string
	WsUrl         string
}

func NewBinanceCrawler(db *database.DB, config BinanceCrawlerConfig) (*BinanceCrawler, error) {

	httpClient := &http.Client{
		Timeout: 90 * time.Second,
	}

	return &BinanceCrawler{
		db:              db,
		baseURL:         config.ApiUrl,
		wsURL:           config.WsUrl,
		httpClient:      httpClient,
		Category:        config.CategoryGUID,
		Ecosystem:       config.EcosystemGUID,
		Period:          config.PeriodGUID,
		defaultLanguage: config.LanguageGUID,
		symbols:         config.Symbols,
		wsStopCh:        make(chan struct{}),
		wsRunning:       false,
	}, nil
}

func (c *BinanceCrawler) SyncPrices(ctx context.Context) error {
	log.Info("Starting Binance price sync", "symbols", c.symbols)

	processedCount := 0
	for _, symbol := range c.symbols {
		apiURL := fmt.Sprintf("%s/ticker/price?symbol=%s", c.baseURL, symbol)

		priceData, err := c.fetchPrice(ctx, apiURL, symbol)
		if err != nil {
			log.Warn("Failed to fetch price for symbol", "symbol", symbol, "err", err)
			continue
		}

		if err := c.processPrice(priceData); err != nil {
			log.Error("Failed to process price", "symbol", symbol, "err", err)
			continue
		}
		processedCount++
	}

	log.Info("Binance price sync completed", "processed", processedCount, "total", len(c.symbols))
	return nil
}

func (c *BinanceCrawler) fetchPrice(ctx context.Context, apiURL, symbol string) (*BinanceTickerPrice, error) {
	retryStrategy := &retry.ExponentialStrategy{
		Min:       3 * time.Second,
		Max:       30 * time.Second,
		MaxJitter: 2 * time.Second,
	}

	var priceData BinanceTickerPrice

	_, err := retry.Do[interface{}](ctx, 3, retryStrategy, func() (interface{}, error) {
		req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Accept", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			log.Warn("API call failed, will retry", "err", err, "url", apiURL)
			return nil, fmt.Errorf("failed to call Binance API: %w", err)
		}
		log.Debug("API call successful", "url", apiURL, "status", resp.StatusCode)
		defer resp.Body.Close()

		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			log.Warn("Failed to read response body, will retry", "err", readErr)
			return nil, fmt.Errorf("failed to read response body: %w", readErr)
		}

		if resp.StatusCode != http.StatusOK {
			if resp.StatusCode >= 400 && resp.StatusCode < 500 {
				return nil, fmt.Errorf("API returned client error %d: %s", resp.StatusCode, string(body))
			}
			log.Warn("API returned server error, will retry", "status", resp.StatusCode, "url", apiURL)
			return nil, fmt.Errorf("API returned server error %d: %s", resp.StatusCode, string(body))
		}

		if err := json.Unmarshal(body, &priceData); err != nil {
			log.Warn("Failed to parse response, will retry", "err", err, "body", string(body))
			return nil, fmt.Errorf("failed to decode API response: %w", err)
		}

		if priceData.Price == "" {
			return nil, fmt.Errorf("price is empty in response")
		}
		if priceData.Symbol == "" {
			priceData.Symbol = symbol
		}
		return nil, nil
	})

	if err == nil {
		return &priceData, nil
	}

	log.Warn("Failed to fetch from endpoint, trying next", "endpoint", apiURL, "err", err)

	return nil, fmt.Errorf("failed to fetch price from all endpoints, last error: %w", err)
}

func (c *BinanceCrawler) processPrice(priceData *BinanceTickerPrice) error {
	return c.db.Transaction(func(txDB *database.DB) error {
		externalID := fmt.Sprintf("BINANCE_%s", priceData.Symbol)

		existingEvent, err := txDB.Event.GetEventByExternalID(externalID)
		if err != nil {
			return fmt.Errorf("failed to check existing event: %w", err)
		}

		price := priceData.Price
		mainScore := "0"
		clusterScore := "0"

		now := time.Now()
		dateCode := now.Format("2006-01-02")
		periodCode := fmt.Sprintf("CRYPTO_BINANCE__%s", dateCode)

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
				Rules:        fmt.Sprintf("Real-time price tracking for %s on Binance", priceData.Symbol),
			}

			if err := txDB.Event.CreateEventLanguage(eventLang); err != nil {
				return fmt.Errorf("failed to create event language: %w", err)
			}

			log.Info("Created Binance event", "symbol", priceData.Symbol, "event_guid", createdEvent.GUID)
		}

		return nil
	})
}

func (c *BinanceCrawler) ensureEventPeriod(txDB *database.DB, periodCode, dateCode string) (string, error) {
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

func (c *BinanceCrawler) updateEventLanguage(txDB *database.DB, eventGUID, title string) error {
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

func (c *BinanceCrawler) StartWebSocketStream(ctx context.Context) error {
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

func (c *BinanceCrawler) StopWebSocketStream() {
	c.wsMutex.Lock()
	defer c.wsMutex.Unlock()

	if !c.wsRunning {
		return
	}

	close(c.wsStopCh)
	if c.wsConn != nil {
		c.wsConn.Close()
		c.wsConn = nil
	}
	c.wsRunning = false
	log.Info("Binance WebSocket stream stopped")
}

func (c *BinanceCrawler) runWebSocketStream(ctx context.Context) {
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

func (c *BinanceCrawler) connectAndSubscribe(ctx context.Context) error {
	dialer := websocket.Dialer{
		HandshakeTimeout: 30 * time.Second,
	}

	streams := make([]string, len(c.symbols))
	for i, symbol := range c.symbols {
		streams[i] = strings.ToLower(symbol) + "@ticker"
	}

	var conn *websocket.Conn
	log.Info("Attempting WebSocket connection", "endpoint", c.wsURL)
	conn, _, err := dialer.Dial(c.wsURL, nil)
	if err != nil {
		log.Warn("WebSocket connection attempt failed", "endpoint", c.wsURL, "err", err)
		return fmt.Errorf("failed to connect to WebSocket endpoint: %w", err)
	}
	log.Info("WebSocket connected successfully", "endpoint", c.wsURL)

	c.wsMutex.Lock()
	c.wsConn = conn
	c.wsMutex.Unlock()

	log.Info("Binance WebSocket connected", "streams", streams)

	subscribeMsg := BinanceWSMessage{
		Method: "SUBSCRIBE",
		Params: make([]interface{}, len(streams)),
		ID:     1,
	}
	for i, stream := range streams {
		subscribeMsg.Params[i] = stream
	}

	if err := conn.WriteJSON(subscribeMsg); err != nil {
		conn.Close()
		return fmt.Errorf("failed to send subscribe message: %w", err)
	}

	var confirmMsg BinanceWSMessage
	if err := conn.ReadJSON(&confirmMsg); err != nil {
		conn.Close()
		return fmt.Errorf("failed to read subscription confirmation: %w", err)
	}

	if confirmMsg.Result != nil {
		log.Info("Binance WebSocket subscribed", "result", confirmMsg.Result)
	}

	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	pingTicker := time.NewTicker(20 * time.Second)
	defer pingTicker.Stop()

	messageCh := make(chan []byte, 100)
	errCh := make(chan error, 1)

	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				errCh <- err
				return
			}
			select {
			case messageCh <- message:
			case <-ctx.Done():
				return
			case <-c.wsStopCh:
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			conn.Close()
			return nil
		case <-c.wsStopCh:
			conn.Close()
			return nil
		case err := <-errCh:
			return fmt.Errorf("WebSocket read error: %w", err)
		case <-pingTicker.C:
			if err := conn.WriteMessage(websocket.PongMessage, []byte{}); err != nil {
				return fmt.Errorf("failed to send pong: %w", err)
			}
		case message := <-messageCh:
			if err := c.handleWebSocketMessage(message); err != nil {
				log.Warn("Failed to handle WebSocket message", "err", err)
			}
		}
	}
}

func (c *BinanceCrawler) handleWebSocketMessage(message []byte) error {
	var tickerStream BinanceTickerStream
	if err := json.Unmarshal(message, &tickerStream); err != nil {
		var wsMsg BinanceWSMessage
		if err2 := json.Unmarshal(message, &wsMsg); err2 == nil {
			return nil
		}
		return fmt.Errorf("failed to unmarshal ticker stream: %w", err)
	}

	if tickerStream.Data.LastPrice == "" {
		return fmt.Errorf("empty price in ticker data")
	}

	priceData := &BinanceTickerPrice{
		Symbol: tickerStream.Data.Symbol,
		Price:  tickerStream.Data.LastPrice,
	}

	if err := c.processPrice(priceData); err != nil {
		return fmt.Errorf("failed to process price update: %w", err)
	}

	log.Debug("Processed real-time price update", "symbol", priceData.Symbol, "price", priceData.Price)
	return nil
}
