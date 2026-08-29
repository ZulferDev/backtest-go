package data

type OHLCV struct {
	Timestamp int64   `parquet:"name=timestamp, type=INT64"`
	Open      float64 `parquet:"name=open, type=DOUBLE"`
	High      float64 `parquet:"name=high, type=DOUBLE"`
	Low       float64 `parquet:"name=low, type=DOUBLE"`
	Close     float64 `parquet:"name=close, type=DOUBLE"`
	Volume    float64 `parquet:"name=volume, type=DOUBLE"`
}
