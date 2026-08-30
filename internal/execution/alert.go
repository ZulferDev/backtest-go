package execution

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// AlertLevel represents the severity of an alert
type AlertLevel string

const (
	AlertLevelInfo     AlertLevel = "INFO"
	AlertLevelWarning  AlertLevel = "WARNING"
	AlertLevelCritical AlertLevel = "CRITICAL"
)

// AlertType represents the type of alert
type AlertType string

const (
	AlertTypeOrderFilled    AlertType = "ORDER_FILLED"
	AlertTypeOrderFailed    AlertType = "ORDER_FAILED"
	AlertTypePositionOpened AlertType = "POSITION_OPENED"
	AlertTypePositionClosed AlertType = "POSITION_CLOSED"
	AlertTypeDrawdown       AlertType = "DRAWDOWN"
	AlertTypeProfit         AlertType = "PROFIT"
	AlertTypeError          AlertType = "ERROR"
	AlertTypeKillSwitch     AlertType = "KILL_SWITCH"
	AlertTypeConnection     AlertType = "CONNECTION"
)

// Alert represents a trading alert
type Alert struct {
	Timestamp int64       `json:"timestamp"`
	Level     AlertLevel  `json:"level"`
	Type      AlertType   `json:"type"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
}

// AlertHandler interface for alert delivery
type AlertHandler interface {
	Send(alert Alert) error
	Name() string
}

// AlertManager manages and routes alerts
type AlertManager struct {
	handlers []AlertHandler
	alerts   []Alert
	mu       sync.RWMutex
	
	// Alert filtering
	minLevel AlertLevel
	
	// File logging
	logFile *os.File
}

// NewAlertManager creates a new alert manager
func NewAlertManager(logPath string, minLevel AlertLevel) (*AlertManager, error) {
	var logFile *os.File
	var err error
	
	if logPath != "" {
		logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open alert log file: %w", err)
		}
	}
	
	return &AlertManager{
		handlers: []AlertHandler{},
		alerts:   []Alert{},
		minLevel: minLevel,
		logFile:  logFile,
	}, nil
}

// AddHandler adds an alert handler
func (am *AlertManager) AddHandler(handler AlertHandler) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.handlers = append(am.handlers, handler)
	log.Printf("Alert handler added: %s", handler.Name())
}

// Send sends an alert through all registered handlers
func (am *AlertManager) Send(alert Alert) {
	// Check if alert meets minimum level
	if !am.meetsMinLevel(alert.Level) {
		return
	}
	
	// Add timestamp if not set
	if alert.Timestamp == 0 {
		alert.Timestamp = time.Now().Unix()
	}
	
	// Store alert
	am.mu.Lock()
	am.alerts = append(am.alerts, alert)
	
	// Write to log file
	if am.logFile != nil {
		alertJSON, _ := json.Marshal(alert)
		_, _ = am.logFile.Write(append(alertJSON, '\n'))
	}
	am.mu.Unlock()
	
	// Send through handlers
	am.mu.RLock()
	handlers := make([]AlertHandler, len(am.handlers))
	copy(handlers, am.handlers)
	am.mu.RUnlock()
	
	for _, handler := range handlers {
		go func(h AlertHandler) {
			if err := h.Send(alert); err != nil {
				log.Printf("Failed to send alert via %s: %v", h.Name(), err)
			}
		}(handler)
	}
	
	// Log to console
	am.logToConsole(alert)
}

// meetsMinLevel checks if alert level meets minimum threshold
func (am *AlertManager) meetsMinLevel(level AlertLevel) bool {
	levelPriority := map[AlertLevel]int{
		AlertLevelInfo:     1,
		AlertLevelWarning:  2,
		AlertLevelCritical: 3,
	}
	
	return levelPriority[level] >= levelPriority[am.minLevel]
}

// logToConsole logs alert to console
func (am *AlertManager) logToConsole(alert Alert) {
	prefix := fmt.Sprintf("[%s][%s]", alert.Level, alert.Type)
	log.Printf("%s %s", prefix, alert.Message)
}

// GetAlerts returns all alerts
func (am *AlertManager) GetAlerts() []Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()
	
	alerts := make([]Alert, len(am.alerts))
	copy(alerts, am.alerts)
	return alerts
}

// GetRecentAlerts returns recent alerts (last N)
func (am *AlertManager) GetRecentAlerts(count int) []Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()
	
	if count > len(am.alerts) {
		count = len(am.alerts)
	}
	
	start := len(am.alerts) - count
	if start < 0 {
		start = 0
	}
	
	alerts := make([]Alert, count)
	copy(alerts, am.alerts[start:])
	return alerts
}

// Close closes the alert manager
func (am *AlertManager) Close() error {
	if am.logFile != nil {
		return am.logFile.Close()
	}
	return nil
}

// ConsoleHandler sends alerts to console (already handled by AlertManager)
type ConsoleHandler struct{}

func NewConsoleHandler() *ConsoleHandler {
	return &ConsoleHandler{}
}

func (h *ConsoleHandler) Name() string {
	return "console"
}

func (h *ConsoleHandler) Send(alert Alert) error {
	// Already logged by AlertManager
	return nil
}

// FileHandler writes alerts to a separate file
type FileHandler struct {
	filePath string
	mu       sync.Mutex
}

func NewFileHandler(filePath string) *FileHandler {
	return &FileHandler{
		filePath: filePath,
	}
}

func (h *FileHandler) Name() string {
	return "file"
}

func (h *FileHandler) Send(alert Alert) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	f, err := os.OpenFile(h.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	
	alertJSON, err := json.Marshal(alert)
	if err != nil {
		return err
	}
	
	_, err = f.Write(append(alertJSON, '\n'))
	return err
}

// WebhookHandler sends alerts via HTTP webhook
type WebhookHandler struct {
	url string
	// Note: Actual HTTP implementation would go here
	// Keeping simple for now to avoid net/http import complexity
}

func NewWebhookHandler(url string) *WebhookHandler {
	return &WebhookHandler{
		url: url,
	}
}

func (h *WebhookHandler) Name() string {
	return "webhook"
}

func (h *WebhookHandler) Send(alert Alert) error {
	// Placeholder: In production, this would POST to webhook URL
	// For now, just log that webhook would be called
	log.Printf("Webhook alert to %s: %s", h.url, alert.Message)
	return nil
}

// Helper functions to create common alerts

// NewOrderAlert creates an order-related alert
func NewOrderAlert(level AlertLevel, orderType AlertType, message string, order *Order) Alert {
	return Alert{
		Level:   level,
		Type:    orderType,
		Message: message,
		Data:    order,
	}
}

// NewPositionAlert creates a position-related alert
func NewPositionAlert(level AlertLevel, posType AlertType, message string, position *Position) Alert {
	return Alert{
		Level:   level,
		Type:    posType,
		Message: message,
		Data:    position,
	}
}

// NewErrorAlert creates an error alert
func NewErrorAlert(message string, err error) Alert {
	return Alert{
		Level:   AlertLevelCritical,
		Type:    AlertTypeError,
		Message: message,
		Data: map[string]interface{}{
			"error": err.Error(),
		},
	}
}

// NewKillSwitchAlert creates a kill switch alert
func NewKillSwitchAlert(reason string, equity float64) Alert {
	return Alert{
		Level:   AlertLevelCritical,
		Type:    AlertTypeKillSwitch,
		Message: fmt.Sprintf("Kill switch triggered: %s", reason),
		Data: map[string]interface{}{
			"reason": reason,
			"equity": equity,
		},
	}
}
