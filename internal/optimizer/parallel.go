package optimizer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ZulferDev/backtest-go/internal/backtest"
	"github.com/ZulferDev/backtest-go/pkg/data"
	"github.com/ZulferDev/backtest-go/pkg/sdk"
)

// StrategyConfig represents a strategy configuration to test
type StrategyConfig struct {
	Name       string
	Strategy   sdk.Strategy
	Parameters map[string]interface{}
}

// BacktestTask represents a single backtest job
type BacktestTask struct {
	ID         string
	Config     StrategyConfig
	Data       []data.OHLCV
	StartTime  time.Time
	EndTime    time.Time
	InitialCap float64
}

// BacktestResult represents the result of a backtest job
type BacktestResult struct {
	TaskID       string
	Config       StrategyConfig
	TotalReturn  float64
	SharpeRatio  float64
	MaxDrawdown  float64
	TotalTrades  int
	WinRate      float64
	ProfitFactor float64
	Duration     time.Duration
	Error        error
}

// ParallelExecutor manages concurrent backtest execution
type ParallelExecutor struct {
	workers      int
	taskQueue    chan BacktestTask
	resultQueue  chan BacktestResult
	wg           sync.WaitGroup
	ctx          context.Context
	cancel       context.CancelFunc
	tasksTotal   int
	tasksRunning int
	mutex        sync.Mutex
}

// NewParallelExecutor creates a new parallel backtest executor
func NewParallelExecutor(workers int) *ParallelExecutor {
	ctx, cancel := context.WithCancel(context.Background())
	return &ParallelExecutor{
		workers:     workers,
		taskQueue:   make(chan BacktestTask, workers*2),
		resultQueue: make(chan BacktestResult, workers*2),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start initializes worker pool
func (p *ParallelExecutor) Start() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

// worker processes backtest tasks
func (p *ParallelExecutor) worker(workerID int) {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			return
		case task, ok := <-p.taskQueue:
			if !ok {
				return
			}

			p.incrementRunning()
			result := p.executeBacktest(task)
			p.decrementRunning()

			select {
			case p.resultQueue <- result:
			case <-p.ctx.Done():
				return
			}
		}
	}
}

// executeBacktest runs a single backtest
func (p *ParallelExecutor) executeBacktest(task BacktestTask) BacktestResult {
	startTime := time.Now()

	result := BacktestResult{
		TaskID: task.ID,
		Config: task.Config,
	}

	// Create backtest engine
	engine := backtest.NewEngine(task.Config.Strategy, task.Data, task.InitialCap)

	// Run backtest
	if err := engine.Run(); err != nil {
		result.Error = fmt.Errorf("backtest execution failed: %w", err)
		result.Duration = time.Since(startTime)
		return result
	}

	// Extract metrics
	state := engine.GetState()
	
	// Calculate total return
	if task.InitialCap > 0 {
		result.TotalReturn = ((state.Equity() - task.InitialCap) / task.InitialCap) * 100
	}

	// Calculate win rate
	trades := state.Trades()
	if len(trades) > 0 {
		winningTrades := 0
		totalProfit := 0.0
		totalLoss := 0.0

		for _, trade := range trades {
			if trade.PnL > 0 {
				winningTrades++
				totalProfit += trade.PnL
			} else if trade.PnL < 0 {
				totalLoss += -trade.PnL
			}
		}

		result.WinRate = (float64(winningTrades) / float64(len(trades))) * 100
		result.TotalTrades = len(trades)

		// Calculate profit factor
		if totalLoss > 0 {
			result.ProfitFactor = totalProfit / totalLoss
		}
	}

	// TODO: Calculate Sharpe and Max Drawdown from equity curve
	// These require equity curve tracking which will be added

	result.Duration = time.Since(startTime)
	return result
}

// Submit adds a backtest task to the queue
func (p *ParallelExecutor) Submit(task BacktestTask) error {
	p.mutex.Lock()
	p.tasksTotal++
	p.mutex.Unlock()

	select {
	case p.taskQueue <- task:
		return nil
	case <-p.ctx.Done():
		return fmt.Errorf("executor is shutting down")
	}
}

// GetResults returns the result channel
func (p *ParallelExecutor) GetResults() <-chan BacktestResult {
	return p.resultQueue
}

// Stop gracefully shuts down the executor
func (p *ParallelExecutor) Stop() {
	close(p.taskQueue)
	p.wg.Wait()
	close(p.resultQueue)
}

// Cancel cancels all running and pending tasks
func (p *ParallelExecutor) Cancel() {
	p.cancel()
	p.Stop()
}

// GetStatus returns current execution status
func (p *ParallelExecutor) GetStatus() (total, running int) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.tasksTotal, p.tasksRunning
}

func (p *ParallelExecutor) incrementRunning() {
	p.mutex.Lock()
	p.tasksRunning++
	p.mutex.Unlock()
}

func (p *ParallelExecutor) decrementRunning() {
	p.mutex.Lock()
	p.tasksRunning--
	p.mutex.Unlock()
}
