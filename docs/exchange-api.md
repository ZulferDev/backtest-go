# Exchange API Reference

## 1. Binance API

### Historical Data (Klines/OHLCV)
- **Endpoint:** `GET https://api.binance.com/api/v3/klines`
- **Rate Limit:** 1200 request weight per minute. Klines weight is 1-2 depending on limit.
- **Parameters:**
  - `symbol` (string): e.g., `BTCUSDT`
  - `interval` (string): `1m`, `3m`, `5m`, `15m`, `30m`, `1h`, `2h`, `4h`, `6h`, `8h`, `12h`, `1d`, `3d`, `1w`, `1M`
  - `limit` (int): default 500, max 1000
  - `startTime` (int): timestamp in ms
  - `endTime` (int): timestamp in ms
- **Response Format:** Array of arrays. Oldest first.
  ```json
  [
    [
      1499040000000,      // Open time
      "0.01634790",       // Open
      "0.80000000",       // High
      "0.01575800",       // Low
      "0.01577100",       // Close
      "148976.11427815",  // Volume
      1499644799999,      // Close time
      "2434.19055334",    // Quote asset volume
      308,                // Number of trades
      "1756.87402397",    // Taker buy base asset volume
      "28.46694368",      // Taker buy quote asset volume
      "17928899.62484339" // Ignore.
    ]
  ]
  ```

---

## 2. Bybit API (V5)

### Historical Data (Kline)
- **Endpoint:** `GET https://api.bybit.com/v5/market/kline`
- **Rate Limit:** 120 req/s per IP.
- **Parameters:**
  - `category` (string): `spot`, `linear`, `inverse`
  - `symbol` (string): e.g., `BTCUSDT`
  - `interval` (string): `1`,`3`,`5`,`15`,`30`,`60`,`120`,`240`,`360`,`720`,`D`,`M`,`W`
  - `limit` (int): default 200, max 1000
  - `start` (int): timestamp in ms
  - `end` (int): timestamp in ms
- **Response Format:** JSON Object containing array of arrays. Newest first.
  ```json
  {
    "retCode": 0,
    "retMsg": "OK",
    "result": {
        "symbol": "BTCUSDT",
        "category": "linear",
        "list": [
            [
                "1670608800000", // startTime
                "17164.0",       // openPrice
                "17164.0",       // highPrice
                "17121.5",       // lowPrice
                "17129.5",       // closePrice
                "612.879",       // volume
                "10499645.1"     // turnover
            ]
        ]
    }
  }
  ```

---

## 3. Data Consistency & Normalization Rules

### Order of Data
- **Binance:** Oldest first (chronological).
- **Bybit:** Newest first (reverse chronological). Needs reversal before storage.

### Types
- Both exchanges return numbers as strings to preserve precision. Must be parsed to `float64`.
- Timestamps are in milliseconds.

### Rate Limiting Strategy
- **Binance:** Use response header `X-MBX-USED-WEIGHT-1M` to track weight. Delay requests if approaching 1000.
- **Bybit:** Relatively high limits (120/s), standard rate limiter is sufficient.

### Recommendation
- Primary Data Source: **Binance** (more history, chronological by default, industry standard).