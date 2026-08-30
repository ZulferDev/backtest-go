package optimizer

import (
	"fmt"
	"sort"
)

// RankingCriteria defines how to rank backtest results
type RankingCriteria struct {
	Metric string  // "return", "sharpe", "profit_factor", "win_rate"
	Weight float64 // 0.0-1.0
}

// ResultAggregator collects and ranks backtest results
type ResultAggregator struct {
	results  []BacktestResult
	criteria []RankingCriteria
}

// NewResultAggregator creates a new result aggregator
func NewResultAggregator(criteria []RankingCriteria) *ResultAggregator {
	return &ResultAggregator{
		results:  make([]BacktestResult, 0),
		criteria: criteria,
	}
}

// Add adds a result to the aggregator
func (r *ResultAggregator) Add(result BacktestResult) {
	r.results = append(r.results, result)
}

// GetAll returns all results
func (r *ResultAggregator) GetAll() []BacktestResult {
	return r.results
}

// GetTopN returns top N results based on ranking criteria
func (r *ResultAggregator) GetTopN(n int) []BacktestResult {
	if len(r.results) == 0 {
		return []BacktestResult{}
	}

	// Calculate scores for all results
	scored := make([]scoredResult, len(r.results))
	for i, result := range r.results {
		scored[i] = scoredResult{
			result: result,
			score:  r.calculateScore(result),
		}
	}

	// Sort by score descending
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Return top N
	if n > len(scored) {
		n = len(scored)
	}

	topResults := make([]BacktestResult, n)
	for i := 0; i < n; i++ {
		topResults[i] = scored[i].result
	}

	return topResults
}

// scoredResult wraps a result with its calculated score
type scoredResult struct {
	result BacktestResult
	score  float64
}

// calculateScore computes a weighted score for a result
func (r *ResultAggregator) calculateScore(result BacktestResult) float64 {
	if result.Error != nil {
		return -1000000.0 // Failed backtests get lowest score
	}

	score := 0.0
	totalWeight := 0.0

	for _, criterion := range r.criteria {
		var metricValue float64

		switch criterion.Metric {
		case "return":
			metricValue = result.TotalReturn
		case "sharpe":
			metricValue = result.SharpeRatio
		case "profit_factor":
			metricValue = result.ProfitFactor
		case "win_rate":
			metricValue = result.WinRate / 100.0 // Normalize to 0-1
		default:
			continue
		}

		score += metricValue * criterion.Weight
		totalWeight += criterion.Weight
	}

	// Normalize by total weight
	if totalWeight > 0 {
		score /= totalWeight
	}

	return score
}

// GetStatistics returns summary statistics of all results
func (r *ResultAggregator) GetStatistics() Statistics {
	if len(r.results) == 0 {
		return Statistics{}
	}

	stats := Statistics{
		Total:      len(r.results),
		Successful: 0,
		Failed:     0,
	}

	var totalReturn, totalSharpe, totalDrawdown, totalWinRate float64
	var totalTrades int

	for _, result := range r.results {
		if result.Error != nil {
			stats.Failed++
			continue
		}

		stats.Successful++
		totalReturn += result.TotalReturn
		totalSharpe += result.SharpeRatio
		totalDrawdown += result.MaxDrawdown
		totalWinRate += result.WinRate
		totalTrades += result.TotalTrades

		if result.TotalReturn > 0 {
			stats.Profitable++
		}
	}

	if stats.Successful > 0 {
		stats.AvgReturn = totalReturn / float64(stats.Successful)
		stats.AvgSharpe = totalSharpe / float64(stats.Successful)
		stats.AvgDrawdown = totalDrawdown / float64(stats.Successful)
		stats.AvgWinRate = totalWinRate / float64(stats.Successful)
		stats.AvgTrades = float64(totalTrades) / float64(stats.Successful)
	}

	return stats
}

// Statistics contains summary statistics
type Statistics struct {
	Total       int
	Successful  int
	Failed      int
	Profitable  int
	AvgReturn   float64
	AvgSharpe   float64
	AvgDrawdown float64
	AvgWinRate  float64
	AvgTrades   float64
}

// GenerateReport creates a text report of results
func (r *ResultAggregator) GenerateReport() string {
	stats := r.GetStatistics()
	top10 := r.GetTopN(10)

	report := "# Mass Optimization Results\n\n"
	report += "## Summary Statistics\n\n"
	report += fmt.Sprintf("- Total Backtests: %d\n", stats.Total)
	report += fmt.Sprintf("- Successful: %d (%.1f%%)\n", stats.Successful, float64(stats.Successful)/float64(stats.Total)*100)
	report += fmt.Sprintf("- Failed: %d\n", stats.Failed)
	report += fmt.Sprintf("- Profitable: %d (%.1f%%)\n", stats.Profitable, float64(stats.Profitable)/float64(stats.Successful)*100)
	report += fmt.Sprintf("- Avg Return: %.2f%%\n", stats.AvgReturn)
	report += fmt.Sprintf("- Avg Sharpe: %.3f\n", stats.AvgSharpe)
	report += fmt.Sprintf("- Avg Win Rate: %.1f%%\n", stats.AvgWinRate)
	report += fmt.Sprintf("- Avg Trades: %.0f\n\n", stats.AvgTrades)

	report += "## Top 10 Strategies\n\n"
	report += "| Rank | Strategy | Return | Sharpe | Win Rate | Trades |\n"
	report += "|------|----------|--------|--------|----------|--------|\n"

	for i, result := range top10 {
		report += fmt.Sprintf("| %d | %s | %.2f%% | %.3f | %.1f%% | %d |\n",
			i+1,
			result.Config.Name,
			result.TotalReturn,
			result.SharpeRatio,
			result.WinRate,
			result.TotalTrades,
		)
	}

	return report
}

// FilterResults returns results matching criteria
func (r *ResultAggregator) FilterResults(filter func(BacktestResult) bool) []BacktestResult {
	filtered := make([]BacktestResult, 0)
	for _, result := range r.results {
		if filter(result) {
			filtered = append(filtered, result)
		}
	}
	return filtered
}

// GetBestByMetric returns the best result for a specific metric
func (r *ResultAggregator) GetBestByMetric(metric string) (BacktestResult, error) {
	if len(r.results) == 0 {
		return BacktestResult{}, fmt.Errorf("no results available")
	}

	best := r.results[0]
	bestValue := r.getMetricValue(best, metric)

	for _, result := range r.results[1:] {
		if result.Error != nil {
			continue
		}

		value := r.getMetricValue(result, metric)
		if value > bestValue {
			best = result
			bestValue = value
		}
	}

	if best.Error != nil {
		return BacktestResult{}, fmt.Errorf("no successful results found")
	}

	return best, nil
}

// getMetricValue extracts a metric value from a result
func (r *ResultAggregator) getMetricValue(result BacktestResult, metric string) float64 {
	switch metric {
	case "return":
		return result.TotalReturn
	case "sharpe":
		return result.SharpeRatio
	case "profit_factor":
		return result.ProfitFactor
	case "win_rate":
		return result.WinRate
	default:
		return 0.0
	}
}
