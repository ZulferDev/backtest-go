# Data Quality Framework

## 1. Validation Rules

- **OHLCV Integrity:** High harus merupakan harga tertinggi di candle tersebut, Low harus terendah. Volume tidak boleh negatif.
- **Completeness:** Rentang waktu bar harus konsisten (selisih timestamp sama dengan durasi bar).
- **Consistency:** Mendeteksi lonjakan harga yang tidak wajar (> 50% dalam 1 bar) untuk mendeteksi data error dari feed exchange.

## 2. Metrics & Thresholds

- **Completeness Score:** % bar yang ada dibanding jumlah bar yang diharapkan. Syarat backtest: > 99.9%.
- **Outlier Threshold:** Perubahan harga > 50% diidentifikasi sebagai anomali kecuali terverifikasi oleh exchange pembanding.