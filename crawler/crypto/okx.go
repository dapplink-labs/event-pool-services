package crypto

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/gorilla/websocket"

	"github.com/multimarket-labs/event-pod-services/database"
)

type OKXCrawler struct {
	db              *database.DB
	Category        string
	Ecosystem       string
	Period          string
	defaultLanguage string
	symbols         []string
	wsConn          *websocket.Conn
	wsMutex         sync.RWMutex
	wsStopCh        chan struct{}
	wsRunning       bool
	wsURL           string
}

type OKXTickerPrice struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

type OKXTickerStream struct {
	Arg struct {
		Channel string `json:"channel"`
		InstID  string `json:"instId"`
	} `json:"arg"`
	Data []struct {
		InstID    string `json:"instId"`
		Last      string `json:"last"`
		LastSz    string `json:"lastSz"`
		AskPx     string `json:"askPx"`
		AskSz     string `json:"askSz"`
		BidPx     string `json:"bidPx"`
		BidSz     string `json:"bidSz"`
		Open24h   string `json:"open24h"`
		High24h   string `json:"high24h"`
		Low24h    string `json:"low24h"`
		Vol24h    string `json:"vol24h"`
		VolCcy24h string `json:"volCcy24h"`
		Ts        string `json:"ts"`
	} `json:"data"`
}

type OKXWSSubscribeMessage struct {
	Op   string     `json:"op"`
	Args []OKXWSArg `json:"args"`
}

type OKXWSArg struct {
	Channel string `json:"channel"`
	InstID  string `json:"instId"`
}

type OKXWSResponse struct {
	Event string `json:"event"`
	Arg   struct {
		Channel string `json:"channel"`
		InstID  string `json:"instId"`
	} `json:"arg"`
	Code string `json:"code"`
	Msg  string `json:"msg"`
}

type OKXCrawlerConfig struct {
	CategoryGUID  string
	EcosystemGUID string
	PeriodGUID    string
	LanguageGUID  string
	Symbols       []string
	WsUrl         string
}

func NewOKXCrawler(db *database.DB, config OKXCrawlerConfig) (*OKXCrawler, error) {
	return &OKXCrawler{
		db:              db,
		Category:        config.CategoryGUID,
		Ecosystem:       config.EcosystemGUID,
		Period:          config.PeriodGUID,
		defaultLanguage: config.LanguageGUID,
		symbols:         config.Symbols,
		wsStopCh:        make(chan struct{}),
		wsRunning:       false,
		wsURL:           config.WsUrl,
	}, nil
}

func (c *OKXCrawler) StartWebSocketStream(ctx context.Context) error {
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

func (c *OKXCrawler) StopWebSocketStream() {
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
	log.Info("OKX WebSocket stream stopped")
}

func (c *OKXCrawler) runWebSocketStream(ctx context.Context) {
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

func (c *OKXCrawler) connectAndSubscribe(ctx context.Context) error {
	if len(c.symbols) == 0 {
		return fmt.Errorf("no symbols configured for OKX crawler")
	}

	//var proxyURL *url.URL
	//proxyStr := os.Getenv("HTTPS_PROXY")
	//if proxyStr == "" {
	//	proxyStr = os.Getenv("HTTP_PROXY")
	//}
	//if proxyStr == "" {
	//	proxyStr = os.Getenv("https_proxy")
	//}
	//if proxyStr == "" {
	//	proxyStr = os.Getenv("http_proxy")
	//}
	//
	//if proxyStr != "" {
	//	var err error
	//	proxyURL, err = url.Parse(proxyStr)
	//	if err != nil {
	//		log.Warn("Failed to parse proxy URL, connecting without proxy", "proxy", proxyStr, "err", err)
	//	} else {
	//		log.Info("OKX WebSocket using proxy", "proxy", proxyStr)
	//	}
	//}

	dialer := websocket.Dialer{
		HandshakeTimeout: 30 * time.Second,
		//Proxy:            http.ProxyURL(proxyURL),
	}

	log.Info("Attempting OKX WebSocket connection", "endpoint", c.wsURL)

	conn, _, err := dialer.Dial(c.wsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to OKX WebSocket: %w", err)
	}

	c.wsMutex.Lock()
	c.wsConn = conn
	c.wsMutex.Unlock()

	log.Info("OKX WebSocket connected successfully")

	args := make([]OKXWSArg, len(c.symbols))
	for i, symbol := range c.symbols {
		instID := c.convertSymbolToOKXFormat(symbol)
		args[i] = OKXWSArg{
			Channel: "tickers",
			InstID:  instID,
		}
	}

	subscribeMsg := OKXWSSubscribeMessage{
		Op:   "subscribe",
		Args: args,
	}

	if err := conn.WriteJSON(subscribeMsg); err != nil {
		conn.Close()
		return fmt.Errorf("failed to send subscribe message: %w", err)
	}

	log.Info("OKX subscription sent", "symbols", c.symbols, "instIds", args)

	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	confirmedCount := 0
	timeout := time.After(10 * time.Second)

	for confirmedCount < len(args) {
		select {
		case <-timeout:
			log.Warn("Subscription confirmation timeout, continuing anyway", "confirmed", confirmedCount, "total", len(args))
			break
		default:
			var confirmMsg OKXWSResponse
			if err := conn.ReadJSON(&confirmMsg); err != nil {
				log.Warn("Failed to read subscription confirmation, continuing anyway", "err", err, "confirmed", confirmedCount)
				break
			}

			if confirmMsg.Event == "subscribe" {
				confirmedCount++
				log.Debug("OKX subscription confirmed", "code", confirmMsg.Code, "msg", confirmMsg.Msg, "arg", confirmMsg.Arg)
			} else if confirmMsg.Event == "error" {
				conn.Close()
				return fmt.Errorf("subscription failed: code=%s, msg=%s", confirmMsg.Code, confirmMsg.Msg)
			} else if confirmMsg.Event == "pong" {
				continue
			}
		}
		if confirmedCount >= len(args) {
			break
		}
	}

	log.Info("OKX WebSocket subscribed successfully", "confirmed", confirmedCount, "total", len(args))

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
			pingMsg := map[string]string{"op": "ping"}
			if err := conn.WriteJSON(pingMsg); err != nil {
				return fmt.Errorf("failed to send ping: %w", err)
			}
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		case message := <-messageCh:
			if err := c.handleWebSocketMessage(message); err != nil {
				log.Warn("Failed to handle WebSocket message", "err", err)
			}
		}
	}
}

func (c *OKXCrawler) convertSymbolToOKXFormat(symbol string) string {
	quoteCurrencies := []string{"USDT", "USDC", "BTC", "ETH", "BNB", "USD"}
	symbolUpper := strings.ToUpper(symbol)

	for _, quote := range quoteCurrencies {
		if strings.HasSuffix(symbolUpper, quote) {
			base := strings.TrimSuffix(symbolUpper, quote)
			return fmt.Sprintf("%s-%s", base, quote)
		}
	}

	return symbolUpper
}

func (c *OKXCrawler) handleWebSocketMessage(message []byte) error {
	var genericMsg map[string]interface{}
	if err := json.Unmarshal(message, &genericMsg); err != nil {
		log.Debug("Failed to parse message as JSON", "err", err, "message", string(message))
		return nil
	}

	if event, ok := genericMsg["event"].(string); ok {
		if event == "pong" || event == "subscribe" || event == "unsubscribe" {
			log.Debug("Received control message", "event", event)
			return nil
		}
		if event == "error" {
			log.Warn("Received error message", "message", string(message))
			return nil
		}
	}

	var tickerStream OKXTickerStream
	if err := json.Unmarshal(message, &tickerStream); err != nil {
		log.Debug("Failed to unmarshal ticker stream", "err", err, "message", string(message))
		return nil
	}

	if tickerStream.Arg.Channel != "tickers" {
		return nil
	}

	if len(tickerStream.Data) == 0 {
		return nil
	}

	tickerData := tickerStream.Data[0]
	lastPrice := tickerData.Last
	if lastPrice == "" {
		log.Debug("Empty price in ticker data", "message", string(message))
		return nil
	}

	instID := tickerData.InstID
	if instID == "" {
		instID = tickerStream.Arg.InstID
	}

	symbol := strings.ReplaceAll(instID, "-", "")

	priceData := &OKXTickerPrice{
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

func (c *OKXCrawler) processPrice(priceData *OKXTickerPrice) error {
	return c.db.Transaction(func(txDB *database.DB) error {
		externalID := fmt.Sprintf("OKX_%s", priceData.Symbol)

		existingEvent, err := txDB.Event.GetEventByExternalID(externalID)
		if err != nil {
			return fmt.Errorf("failed to check existing event: %w", err)
		}

		price := priceData.Price
		mainScore := "0"
		clusterScore := "0"

		now := time.Now()
		dateCode := now.Format("2006-01-02")
		periodCode := fmt.Sprintf("CRYPTO_%s", dateCode)

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
				Rules:        fmt.Sprintf("Real-time price tracking for %s on OKX", priceData.Symbol),
			}

			if err := txDB.Event.CreateEventLanguage(eventLang); err != nil {
				return fmt.Errorf("failed to create event language: %w", err)
			}

			log.Info("Created OKX event", "symbol", priceData.Symbol, "event_guid", createdEvent.GUID)
		}

		return nil
	})
}

func (c *OKXCrawler) ensureEventPeriod(txDB *database.DB, periodCode, dateCode string) (string, error) {
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

func (c *OKXCrawler) updateEventLanguage(txDB *database.DB, eventGUID, title string) error {
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
