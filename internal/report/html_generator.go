package report

import (
	"fmt"
	"html/template"
	"os"
	"time"

	"github.com/ZulferDev/backtest-go/internal/backtest"
	"github.com/ZulferDev/backtest-go/internal/metrics"
)

// HTMLGenerator generates HTML reports from backtest results
type HTMLGenerator struct {
	templatePath string
}

// NewHTMLGenerator creates a new HTML report generator
func NewHTMLGenerator() *HTMLGenerator {
	return &HTMLGenerator{}
}

// ReportData holds data for HTML report generation
type ReportData struct {
	Title           string
	GeneratedAt     string
	Strategy        string
	
	// Summary metrics
	InitialCapital  float64
	FinalEquity     float64
	TotalReturn     float64
	TotalReturnPct  float64
	
	TotalTrades     int
	WinningTrades   int
	LosingTrades    int
	WinRate         float64
	
	SharpeRatio     float64
	SortinoRatio    float64
	MaxDrawdown     float64
	ProfitFactor    float64
	
	AverageWin      float64
	AverageLoss     float64
	WinLossRatio    float64
	
	// Trade history
	Trades          []TradeRow
	
	// Equity curve data
	EquityTimestamps []string
	EquityValues     []float64
}

// TradeRow represents a single trade in the report
type TradeRow struct {
	Number     int
	EntryTime  string
	ExitTime   string
	Side       string
	EntryPrice float64
	ExitPrice  float64
	Size       float64
	PnL        float64
	PnLPct     float64
	Fee        float64
}

// Generate creates an HTML report from backtest state
func (g *HTMLGenerator) Generate(state *backtest.State, strategyName string, outputPath string) error {
	trades := state.Trades()
	initialCash := state.InitialCash()
	finalEquity := state.Equity()
	
	// Calculate metrics
	calc := metrics.NewCalculator(trades, initialCash)
	
	// Build report data
	data := ReportData{
		Title:          "Backtest Report",
		GeneratedAt:    time.Now().Format("2006-01-02 15:04:05"),
		Strategy:       strategyName,
		InitialCapital: initialCash,
		FinalEquity:    finalEquity,
		TotalReturn:    finalEquity - initialCash,
		TotalReturnPct: calc.TotalReturn(finalEquity) * 100,
		TotalTrades:    len(trades),
		SharpeRatio:    calc.SharpeRatio(),
		SortinoRatio:   calc.SortinoRatio(),
		MaxDrawdown:    calc.MaxDrawdown(),
		ProfitFactor:   calc.ProfitFactor(),
		WinRate:        calc.WinRate(),
	}
	
	// Process trades
	wins := 0
	losses := 0
	var totalWinPnL, totalLossPnL float64
	
	for i, trade := range trades {
		if trade.PnL > 0 {
			wins++
			totalWinPnL += trade.PnL
		} else {
			losses++
			totalLossPnL += -trade.PnL
		}
		
		pnlPct := 0.0
		if trade.EntryPrice > 0 {
			pnlPct = (trade.ExitPrice - trade.EntryPrice) / trade.EntryPrice * 100
			if trade.Side == "short" {
				pnlPct = -pnlPct
			}
		}
		
		data.Trades = append(data.Trades, TradeRow{
			Number:     i + 1,
			EntryTime:  time.Unix(trade.EntryTime, 0).Format("2006-01-02 15:04"),
			ExitTime:   time.Unix(trade.ExitTime, 0).Format("2006-01-02 15:04"),
			Side:       trade.Side,
			EntryPrice: trade.EntryPrice,
			ExitPrice:  trade.ExitPrice,
			Size:       trade.Size,
			PnL:        trade.PnL,
			PnLPct:     pnlPct,
			Fee:        trade.Fee,
		})
	}
	
	data.WinningTrades = wins
	data.LosingTrades = losses
	
	if wins > 0 {
		data.AverageWin = totalWinPnL / float64(wins)
	}
	if losses > 0 {
		data.AverageLoss = totalLossPnL / float64(losses)
	}
	if data.AverageLoss > 0 {
		data.WinLossRatio = data.AverageWin / data.AverageLoss
	}
	
	// Build equity curve
	equity := initialCash
	data.EquityTimestamps = append(data.EquityTimestamps, "Start")
	data.EquityValues = append(data.EquityValues, equity)
	
	for _, trade := range trades {
		equity += trade.PnL
		data.EquityTimestamps = append(data.EquityTimestamps, 
			time.Unix(trade.ExitTime, 0).Format("2006-01-02"))
		data.EquityValues = append(data.EquityValues, equity)
	}
	
	// Generate HTML
	tmpl := g.getTemplate()
	
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()
	
	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}
	
	return nil
}

// getTemplate returns the HTML template
func (g *HTMLGenerator) getTemplate() *template.Template {
	tmplStr := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background: #f5f7fa; padding: 20px; }
        .container { max-width: 1400px; margin: 0 auto; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; border-radius: 10px; margin-bottom: 30px; }
        .header h1 { font-size: 32px; margin-bottom: 10px; }
        .header p { opacity: 0.9; }
        .metrics-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 20px; margin-bottom: 30px; }
        .metric-card { background: white; padding: 20px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .metric-label { font-size: 14px; color: #666; margin-bottom: 8px; }
        .metric-value { font-size: 28px; font-weight: bold; color: #333; }
        .metric-value.positive { color: #10b981; }
        .metric-value.negative { color: #ef4444; }
        .section { background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); margin-bottom: 30px; }
        .section h2 { font-size: 24px; margin-bottom: 20px; color: #333; border-bottom: 2px solid #667eea; padding-bottom: 10px; }
        table { width: 100%; border-collapse: collapse; }
        th, td { padding: 12px; text-align: left; border-bottom: 1px solid #eee; }
        th { background: #f8f9fa; font-weight: 600; color: #333; }
        tr:hover { background: #f8f9fa; }
        .long { color: #10b981; font-weight: 600; }
        .short { color: #ef4444; font-weight: 600; }
        .chart { height: 400px; margin: 20px 0; }
    </style>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.Title}}</h1>
            <p>Strategy: {{.Strategy}} | Generated: {{.GeneratedAt}}</p>
        </div>

        <div class="metrics-grid">
            <div class="metric-card">
                <div class="metric-label">Initial Capital</div>
                <div class="metric-value">${{printf "%.2f" .InitialCapital}}</div>
            </div>
            <div class="metric-card">
                <div class="metric-label">Final Equity</div>
                <div class="metric-value">${{printf "%.2f" .FinalEquity}}</div>
            </div>
            <div class="metric-card">
                <div class="metric-label">Total Return</div>
                <div class="metric-value {{if ge .TotalReturn 0.0}}positive{{else}}negative{{end}}">
                    {{printf "%.2f%%" .TotalReturnPct}}
                </div>
            </div>
            <div class="metric-card">
                <div class="metric-label">Total Trades</div>
                <div class="metric-value">{{.TotalTrades}}</div>
            </div>
            <div class="metric-card">
                <div class="metric-label">Win Rate</div>
                <div class="metric-value">{{printf "%.1f%%" .WinRate}}</div>
            </div>
            <div class="metric-card">
                <div class="metric-label">Sharpe Ratio</div>
                <div class="metric-value">{{printf "%.2f" .SharpeRatio}}</div>
            </div>
            <div class="metric-card">
                <div class="metric-label">Max Drawdown</div>
                <div class="metric-value negative">{{printf "%.2f%%" .MaxDrawdown}}</div>
            </div>
            <div class="metric-card">
                <div class="metric-label">Profit Factor</div>
                <div class="metric-value">{{printf "%.2f" .ProfitFactor}}</div>
            </div>
        </div>

        <div class="section">
            <h2>Equity Curve</h2>
            <canvas id="equityChart" class="chart"></canvas>
        </div>

        <div class="section">
            <h2>Trade History</h2>
            <table>
                <thead>
                    <tr>
                        <th>#</th>
                        <th>Entry Time</th>
                        <th>Exit Time</th>
                        <th>Side</th>
                        <th>Entry Price</th>
                        <th>Exit Price</th>
                        <th>Size</th>
                        <th>PnL</th>
                        <th>PnL %</th>
                        <th>Fee</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .Trades}}
                    <tr>
                        <td>{{.Number}}</td>
                        <td>{{.EntryTime}}</td>
                        <td>{{.ExitTime}}</td>
                        <td class="{{.Side}}">{{.Side}}</td>
                        <td>{{printf "%.2f" .EntryPrice}}</td>
                        <td>{{printf "%.2f" .ExitPrice}}</td>
                        <td>{{printf "%.4f" .Size}}</td>
                        <td class="{{if ge .PnL 0.0}}long{{else}}short{{end}}">{{printf "%.2f" .PnL}}</td>
                        <td>{{printf "%.2f%%" .PnLPct}}</td>
                        <td>{{printf "%.2f" .Fee}}</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </div>
    </div>

    <script>
        const ctx = document.getElementById('equityChart').getContext('2d');
        new Chart(ctx, {
            type: 'line',
            data: {
                labels: [{{range $i, $v := .EquityTimestamps}}{{if $i}},{{end}}"{{$v}}"{{end}}],
                datasets: [{
                    label: 'Equity',
                    data: [{{range $i, $v := .EquityValues}}{{if $i}},{{end}}{{$v}}{{end}}],
                    borderColor: '#667eea',
                    backgroundColor: 'rgba(102, 126, 234, 0.1)',
                    borderWidth: 2,
                    fill: true,
                    tension: 0.4
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: { display: false },
                    tooltip: {
                        callbacks: {
                            label: function(context) {
                                return 'Equity: $' + context.parsed.y.toFixed(2);
                            }
                        }
                    }
                },
                scales: {
                    y: {
                        beginAtZero: false,
                        ticks: {
                            callback: function(value) {
                                return '$' + value.toFixed(0);
                            }
                        }
                    }
                }
            }
        });
    </script>
</body>
</html>`

	return template.Must(template.New("report").Parse(tmplStr))
}
