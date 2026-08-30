package execution

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// MockExchangeAdapter implements ExchangeAdapter for testing
type MockExchangeAdapter struct {
	connected   bool
	balance     float64
	position    *Position
	orders      map[string]*Order
	orderIDSeq  int
	
	// Simulate behaviors
	shouldFailConnect bool
	shouldFailOrder   bool
	orderFillDelay    time.Duration
}

func NewMockExchangeAdapter() *MockExchangeAdapter {
	return &MockExchangeAdapter{
		connected:      false,
		balance:        10000.0,
		orders:         make(map[string]*Order),
		orderFillDelay: 100 * time.Millisecond,
	}
}

func (m *MockExchangeAdapter) Connect(ctx context.Context) error {
	if m.shouldFailConnect {
		return fmt.Errorf("mock connection failed")
	}
	m.connected = true
	return nil
}

func (m *MockExchangeAdapter) Disconnect() error {
	m.connected = false
	return nil
}

func (m *MockExchangeAdapter) IsConnected() bool {
	return m.connected
}

func (m *MockExchangeAdapter) PlaceOrder(ctx context.Context, symbol string, side OrderSide, orderType OrderType, quantity float64, price float64) (*Order, error) {
	if m.shouldFailOrder {
		return nil, fmt.Errorf("mock order placement failed")
	}
	
	m.orderIDSeq++
	orderID := fmt.Sprintf("ORDER_%d", m.orderIDSeq)
	
	order := &Order{
		ID:         orderID,
		Symbol:     symbol,
		Side:       side,
		Type:       orderType,
		Quantity:   quantity,
		Price:      price,
		Status:     OrderStatusPending,
		Timestamp:  time.Now().Unix(),
		ExchangeID: "MOCK_" + orderID,
	}
	
	m.orders[orderID] = order
	
	// Simulate async fill
	go func() {
		time.Sleep(m.orderFillDelay)
		order.Status = OrderStatusFilled
		order.FilledQty = quantity
		order.AvgFillPrice = 50000.0 // Mock price
		
		// Update mock position
		if side == OrderSideBuy {
			m.position = &Position{
				Symbol:     symbol,
				Side:       "long",
				Size:       quantity,
				EntryPrice: order.AvgFillPrice,
				EntryTime:  time.Now().Unix(),
			}
		} else {
			m.position = nil // Close position
		}
	}()
	
	return order, nil
}

func (m *MockExchangeAdapter) CancelOrder(ctx context.Context, orderID string) error {
	order, exists := m.orders[orderID]
	if !exists {
		return fmt.Errorf("order not found")
	}
	order.Status = OrderStatusCancelled
	return nil
}

func (m *MockExchangeAdapter) GetOrder(ctx context.Context, orderID string) (*Order, error) {
	order, exists := m.orders[orderID]
	if !exists {
		return nil, fmt.Errorf("order not found")
	}
	return order, nil
}

func (m *MockExchangeAdapter) GetBalance(ctx context.Context) (float64, error) {
	return m.balance, nil
}

func (m *MockExchangeAdapter) GetPosition(ctx context.Context, symbol string) (*Position, error) {
	if m.position == nil {
		return nil, fmt.Errorf("no position")
	}
	return m.position, nil
}

func (m *MockExchangeAdapter) GetCurrentPrice(ctx context.Context, symbol string) (float64, error) {
	return 50000.0, nil
}

// Tests

func TestBridgeCreation(t *testing.T) {
	ctx := context.Background()
	mockAdapter := NewMockExchangeAdapter()
	
	bridge, err := NewBridge(ctx, BridgeConfig{
		Adapter:     mockAdapter,
		Strategy:    nil, // Would need mock strategy
		Symbol:      "BTCUSDT",
		InitialCash: 10000,
		MaxDrawdown: 0.10,
	})
	
	if err == nil {
		t.Error("Expected error for nil strategy")
	}
	
	// Test with valid config would need mock strategy
	_ = bridge
}

func TestKillSwitch(t *testing.T) {
	config := KillSwitchConfig{
		MaxDrawdown:     0.10,
		MaxDailyLoss:    1000,
		MaxPositionSize: 1.0,
	}
	
	ks := NewKillSwitch(config)
	
	// Test no trigger at start
	shouldKill, reason := ks.Evaluate(10000, 10000, nil)
	if shouldKill {
		t.Errorf("Should not trigger at start: %s", reason)
	}
	
	// Test drawdown trigger
	ks.peakEquity = 10000
	shouldKill, reason = ks.Evaluate(8500, 10000, nil)
	if !shouldKill {
		t.Error("Should trigger on 15% drawdown")
	}
	if reason == "" {
		t.Error("Expected reason for trigger")
	}
	
	// Test position size trigger
	ks2 := NewKillSwitch(config)
	largePos := &Position{
		Size: 2.0,
		Side: "long",
	}
	shouldKill, reason = ks2.Evaluate(10000, 10000, largePos)
	if !shouldKill {
		t.Error("Should trigger on oversized position")
	}
}

func TestAlertManager(t *testing.T) {
	am, err := NewAlertManager("", AlertLevelInfo)
	if err != nil {
		t.Fatalf("Failed to create alert manager: %v", err)
	}
	defer am.Close()
	
	// Add console handler
	am.AddHandler(NewConsoleHandler())
	
	// Send test alert
	alert := Alert{
		Level:   AlertLevelWarning,
		Type:    AlertTypeDrawdown,
		Message: "Test alert",
	}
	
	am.Send(alert)
	
	// Check alert was stored
	alerts := am.GetAlerts()
	if len(alerts) != 1 {
		t.Errorf("Expected 1 alert, got %d", len(alerts))
	}
	
	if alerts[0].Message != "Test alert" {
		t.Errorf("Alert message mismatch: %s", alerts[0].Message)
	}
}

func TestAlertFiltering(t *testing.T) {
	am, _ := NewAlertManager("", AlertLevelWarning)
	defer am.Close()
	
	// Send info alert (should be filtered)
	am.Send(Alert{
		Level:   AlertLevelInfo,
		Type:    AlertTypeConnection,
		Message: "Info message",
	})
	
	// Send warning alert (should pass)
	am.Send(Alert{
		Level:   AlertLevelWarning,
		Type:    AlertTypeDrawdown,
		Message: "Warning message",
	})
	
	alerts := am.GetAlerts()
	if len(alerts) != 1 {
		t.Errorf("Expected 1 alert after filtering, got %d", len(alerts))
	}
	
	if alerts[0].Level != AlertLevelWarning {
		t.Error("Wrong alert passed through filter")
	}
}

func TestKillSwitchReset(t *testing.T) {
	config := KillSwitchConfig{
		MaxDrawdown: 0.10,
	}
	
	ks := NewKillSwitch(config)
	ks.peakEquity = 10000
	ks.dailyTradeCount = 5
	ks.consecutiveLosses = 3
	
	// Reset
	ks.Reset(9000)
	
	if ks.peakEquity != 9000 {
		t.Errorf("Peak equity not reset: %.2f", ks.peakEquity)
	}
	
	if ks.dailyTradeCount != 0 {
		t.Error("Daily trade count not reset")
	}
	
	if ks.consecutiveLosses != 0 {
		t.Error("Consecutive losses not reset")
	}
}

func TestKillSwitchTradeRecording(t *testing.T) {
	config := KillSwitchConfig{
		MaxConsecutiveLosses: 3,
	}
	
	ks := NewKillSwitch(config)
	
	// Record winning trade
	ks.RecordTrade(100)
	if ks.consecutiveLosses != 0 {
		t.Error("Consecutive losses should reset on win")
	}
	
	// Record losing trades
	ks.RecordTrade(-50)
	ks.RecordTrade(-50)
	ks.RecordTrade(-50)
	
	if ks.consecutiveLosses != 3 {
		t.Errorf("Expected 3 consecutive losses, got %d", ks.consecutiveLosses)
	}
	
	// Should trigger on next evaluation
	shouldKill, reason := ks.Evaluate(9500, 10000, nil)
	if !shouldKill {
		t.Error("Should trigger after max consecutive losses")
	}
	if reason == "" {
		t.Error("Expected reason for trigger")
	}
}
