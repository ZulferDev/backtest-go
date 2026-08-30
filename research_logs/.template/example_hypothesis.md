# Hypothesis: RSI Mean Reversion

## Market Observation
- Cryptocurrency markets (especially BTC) exhibit mean-reverting behavior during ranging periods
- RSI oversold conditions (< 30) often precede short-term price bounces
- RSI overbought conditions (> 70) often precede short-term pullbacks
- This pattern is more reliable during low-volatility sideways markets

## Core Thesis
When RSI reaches extreme levels (oversold/overbought), price tends to revert to the mean. By entering positions at RSI extremes and exiting near neutral levels, we can capture mean reversion profits with favorable risk/reward.

## Expected Edge
- **Entry**: 
  - Buy when RSI(14) crosses below 30 (oversold)
  - Sell/Short when RSI(14) crosses above 70 (overbought)
- **Exit**: 
  - Close long when RSI rises back above 50 (neutral)
  - Close short when RSI falls back below 50 (neutral)
- **Risk Management**: 
  - Fixed 5% stop loss from entry
  - Position size: 95% of equity per trade
  - No pyramiding (one position at a time)

## Success Criteria
- Sharpe Ratio > 1.5
- Max Drawdown < 15%
- Win Rate > 45%
- Total Return > 20% (annualized)

## Implementation Notes
- Timeframe: 1h (hourly bars)
- Symbols: BTCUSDT
- Indicators needed: RSI(14)
- Position sizing: Fixed fractional (95% equity)

## Expected Challenges
- May fail in strong trending markets (consecutive stop-outs)
- RSI can stay oversold/overbought for extended periods during trends
- Need sufficient historical data (50+ bars for RSI warmup)
- Performance highly dependent on volatility regime
