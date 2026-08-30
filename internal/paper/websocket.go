package paper

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ZulferDev/backtest-go/pkg/data"
)

// WebSocketClient handles real-time market data streaming
type WebSocketClient struct {
	conn         *websocket.Conn
	symbol       string
	interval     string
	subscribers  []chan data.OHLCV
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	reconnecting bool
}

// BinanceKlineMessage represents Binance WebSocket kline/candlestick event
type BinanceKlineMessage struct {
	EventType string `json:"e"`
	EventTime int64  `json:"E"`
	Symbol    string `json:"s"`
	Kline     struct {
		StartTime            int64  `json:"t"`
		CloseTime            int64  `json:"T"`
		Symbol               string `json:"s"`
		Interval             string `json:"i"`
		Open                 string `json:"o"`
		Close                string `json:"c"`
		High                 string `json:"h"`
		Low                  string `json:"l"`
		Volume               string `json:"v"`
		IsClosed             bool   `json:"x"`
		QuoteAssetVolume     string `json:"q"`
		NumberOfTrades       int    `json:"n"`
		TakerBuyBaseVolume   string `json:"V"`
		TakerBuyQuoteVolume  string `json:"Q"`
	} `json:"k"`
}

// NewWebSocketClient creates a new WebSocket client for market data
func NewWebSocketClient(ctx context.Context, symbol, interval string) (*WebSocketClient, error) {
	clientCtx, cancel := context.WithCancel(ctx)
	
	wsc := &WebSocketClient{
		symbol:      symbol,
		interval:    interval,
		subscribers: []chan data.OHLCV{},
		ctx:         clientCtx,
		cancel:      cancel,
	}

	if err := wsc.connect(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect websocket: %w", err)
	}

	go wsc.readLoop()
	go wsc.heartbeat()

	return wsc, nil
}

// connect establishes WebSocket connection to Binance
func (wsc *WebSocketClient) connect() error {
	streamName := fmt.Sprintf("%s@kline_%s", wsc.symbol, wsc.interval)
	wsURL := fmt.Sprintf("wss://stream.binance.com:9443/ws/%s", streamName)

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("websocket dial failed: %w", err)
	}

	wsc.conn = conn
	log.Printf("WebSocket connected to %s", wsURL)
	return nil
}

// reconnect attempts to reconnect with exponential backoff
func (wsc *WebSocketClient) reconnect() {
	if wsc.reconnecting {
		return
	}
	
	wsc.reconnecting = true
	defer func() { wsc.reconnecting = false }()

	for attempt := 0; attempt < 5; attempt++ {
		select {
		case <-wsc.ctx.Done():
			return
		case <-time.After(time.Duration(1<<attempt) * time.Second):
		}

		log.Printf("Reconnecting WebSocket (attempt %d/5)...", attempt+1)
		if err := wsc.connect(); err != nil {
			log.Printf("Reconnect failed: %v", err)
			continue
		}

		log.Println("WebSocket reconnected successfully")
		return
	}

	log.Println("WebSocket reconnection failed after 5 attempts")
	wsc.cancel()
}

// readLoop continuously reads messages from WebSocket
func (wsc *WebSocketClient) readLoop() {
	defer wsc.Close()

	for {
		select {
		case <-wsc.ctx.Done():
			return
		default:
		}

		var msg BinanceKlineMessage
		if err := wsc.conn.ReadJSON(&msg); err != nil {
			log.Printf("WebSocket read error: %v", err)
			wsc.reconnect()
			continue
		}

		// Only process closed candles for consistency with backtest
		if !msg.Kline.IsClosed {
			continue
		}

		ohlcv := wsc.parseKline(msg)
		wsc.broadcast(ohlcv)
	}
}

// parseKline converts Binance kline message to OHLCV
func (wsc *WebSocketClient) parseKline(msg BinanceKlineMessage) data.OHLCV {
	return data.OHLCV{
		Timestamp: msg.Kline.StartTime,
		Open:      parseFloat(msg.Kline.Open),
		High:      parseFloat(msg.Kline.High),
		Low:       parseFloat(msg.Kline.Low),
		Close:     parseFloat(msg.Kline.Close),
		Volume:    parseFloat(msg.Kline.Volume),
	}
}

// broadcast sends OHLCV data to all subscribers
func (wsc *WebSocketClient) broadcast(ohlcv data.OHLCV) {
	wsc.mu.RLock()
	defer wsc.mu.RUnlock()

	for _, ch := range wsc.subscribers {
		select {
		case ch <- ohlcv:
		default:
			log.Println("Warning: subscriber channel full, dropping candle")
		}
	}
}

// Subscribe returns a channel that receives OHLCV data
func (wsc *WebSocketClient) Subscribe() chan data.OHLCV {
	wsc.mu.Lock()
	defer wsc.mu.Unlock()

	ch := make(chan data.OHLCV, 100)
	wsc.subscribers = append(wsc.subscribers, ch)
	return ch
}

// heartbeat sends ping messages to keep connection alive
func (wsc *WebSocketClient) heartbeat() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-wsc.ctx.Done():
			return
		case <-ticker.C:
			if wsc.conn != nil {
				if err := wsc.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
					log.Printf("Ping error: %v", err)
				}
			}
		}
	}
}

// Close closes the WebSocket connection and all subscriber channels
func (wsc *WebSocketClient) Close() error {
	wsc.cancel()

	wsc.mu.Lock()
	defer wsc.mu.Unlock()

	for _, ch := range wsc.subscribers {
		close(ch)
	}
	wsc.subscribers = nil

	if wsc.conn != nil {
		return wsc.conn.Close()
	}
	return nil
}

// parseFloat safely parses string to float64
func parseFloat(s string) float64 {
	var f float64
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		log.Printf("Failed to parse float: %v", err)
	}
	return f
}
