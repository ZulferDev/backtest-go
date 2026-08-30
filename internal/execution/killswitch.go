package execution

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// KillSwitchConfig holds configuration for kill switch
type KillSwitchConfig struct {
	MaxDrawdown     float64 // Maximum drawdown allowed (e.g., 0.10 for 10%)
	MaxDailyLoss    float64 // Maximum daily loss allowed (absolute value)
	MaxPositionSize float64 // Maximum position size allowed
	
	// Time-based limits
	MaxDailyTrades  int     // Maximum trades per day
	MaxConsecutiveLosses int // Maximum consecutive losses before stop
}

// KillSwitch implements safety limits for live trading
type KillSwitch struct {
	config KillSwitchConfig
	
	// State tracking
	peakEquity        float64
	dailyStartEquity  float64
	dailyTradeCount   int
	consecutiveLosses int
	lastResetDate     string
	
	mu sync.RWMutex
	
	// Trigger history
	triggerHistory []TriggerEvent
}

// TriggerEvent records a kill switch trigger
type TriggerEvent struct {
	Timestamp int64
	Reason    string
	Equity    float64
	Drawdown  float64
}

// NewKillSwitch creates a new kill switch with given config
func NewKillSwitch(config KillSwitchConfig) *KillSwitch {
	return &KillSwitch{
		config:           config,
		peakEquity:       0,
		dailyStartEquity: 0,
		dailyTradeCount:  0,
		consecutiveLosses: 0,
		lastResetDate:    currentDate(),
		triggerHistory:   []TriggerEvent{},
	}
}

// Evaluate checks if any kill switch condition is met
func (ks *KillSwitch) Evaluate(currentEquity, initialEquity float64, position *Position) (bool, string) {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	
	// Reset daily counters if new day
	if currentDate() != ks.lastResetDate {
		ks.resetDaily(currentEquity)
	}
	
	// Update peak equity
	if currentEquity > ks.peakEquity {
		ks.peakEquity = currentEquity
	}
	
	// Check max drawdown from peak
	if ks.config.MaxDrawdown > 0 && ks.peakEquity > 0 {
		drawdown := (ks.peakEquity - currentEquity) / ks.peakEquity
		if drawdown > ks.config.MaxDrawdown {
			reason := fmt.Sprintf("Max drawdown exceeded: %.2f%% (limit: %.2f%%)", 
				drawdown*100, ks.config.MaxDrawdown*100)
			ks.recordTrigger(reason, currentEquity, drawdown)
			return true, reason
		}
	}
	
	// Check max daily loss
	if ks.config.MaxDailyLoss > 0 {
		dailyLoss := ks.dailyStartEquity - currentEquity
		if dailyLoss > ks.config.MaxDailyLoss {
			reason := fmt.Sprintf("Max daily loss exceeded: %.2f (limit: %.2f)", 
				dailyLoss, ks.config.MaxDailyLoss)
			ks.recordTrigger(reason, currentEquity, 0)
			return true, reason
		}
	}
	
	// Check max position size
	if ks.config.MaxPositionSize > 0 && position != nil {
		if position.Size > ks.config.MaxPositionSize {
			reason := fmt.Sprintf("Position size exceeded: %.4f (limit: %.4f)", 
				position.Size, ks.config.MaxPositionSize)
			ks.recordTrigger(reason, currentEquity, 0)
			return true, reason
		}
	}
	
	// Check max daily trades
	if ks.config.MaxDailyTrades > 0 && ks.dailyTradeCount >= ks.config.MaxDailyTrades {
		reason := fmt.Sprintf("Max daily trades exceeded: %d (limit: %d)", 
			ks.dailyTradeCount, ks.config.MaxDailyTrades)
		ks.recordTrigger(reason, currentEquity, 0)
		return true, reason
	}
	
	// Check consecutive losses
	if ks.config.MaxConsecutiveLosses > 0 && ks.consecutiveLosses >= ks.config.MaxConsecutiveLosses {
		reason := fmt.Sprintf("Max consecutive losses exceeded: %d (limit: %d)", 
			ks.consecutiveLosses, ks.config.MaxConsecutiveLosses)
		ks.recordTrigger(reason, currentEquity, 0)
		return true, reason
	}
	
	return false, ""
}

// RecordTrade records a trade for kill switch tracking
func (ks *KillSwitch) RecordTrade(pnl float64) {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	
	ks.dailyTradeCount++
	
	if pnl < 0 {
		ks.consecutiveLosses++
	} else {
		ks.consecutiveLosses = 0
	}
}

// resetDaily resets daily counters
func (ks *KillSwitch) resetDaily(currentEquity float64) {
	ks.dailyStartEquity = currentEquity
	ks.dailyTradeCount = 0
	ks.lastResetDate = currentDate()
	log.Printf("Kill switch daily reset - Starting equity: %.2f", currentEquity)
}

// recordTrigger records a kill switch trigger event
func (ks *KillSwitch) recordTrigger(reason string, equity, drawdown float64) {
	event := TriggerEvent{
		Timestamp: time.Now().Unix(),
		Reason:    reason,
		Equity:    equity,
		Drawdown:  drawdown,
	}
	ks.triggerHistory = append(ks.triggerHistory, event)
	log.Printf("KILL SWITCH TRIGGERED: %s", reason)
}

// GetTriggerHistory returns the trigger history
func (ks *KillSwitch) GetTriggerHistory() []TriggerEvent {
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	
	history := make([]TriggerEvent, len(ks.triggerHistory))
	copy(history, ks.triggerHistory)
	return history
}

// GetStats returns current kill switch stats
func (ks *KillSwitch) GetStats() map[string]interface{} {
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	
	return map[string]interface{}{
		"peak_equity":         ks.peakEquity,
		"daily_start_equity":  ks.dailyStartEquity,
		"daily_trade_count":   ks.dailyTradeCount,
		"consecutive_losses":  ks.consecutiveLosses,
		"last_reset_date":     ks.lastResetDate,
		"trigger_count":       len(ks.triggerHistory),
	}
}

// Reset resets the kill switch state (use with caution)
func (ks *KillSwitch) Reset(currentEquity float64) {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	
	ks.peakEquity = currentEquity
	ks.dailyStartEquity = currentEquity
	ks.dailyTradeCount = 0
	ks.consecutiveLosses = 0
	ks.lastResetDate = currentDate()
	
	log.Printf("Kill switch manually reset - Equity: %.2f", currentEquity)
}

// currentDate returns current date in YYYY-MM-DD format
func currentDate() string {
	return time.Now().Format("2006-01-02")
}
