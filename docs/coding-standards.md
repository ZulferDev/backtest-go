# Coding Standards & Testing Guidelines

## 1. Go Style Guide (Core Framework)

- Ikuti standard `gofmt` dan `goimports`.
- Gunakan `golangci-lint` sebagai static analysis checker.
- Hindari penggunaan global state atau package-level mutable variables.
- Performa sangat kritikal di core engine: hindari alokasi heap di dalam hot path (event loop `OnBar`). Gunakan pre-allocated slice atau object pooling.

## 2. Standards untuk "AI-Generated Strategy Code"

AI diperlakukan sebagai kontributor kode. Namun, kode AI memiliki aturan khusus (Sandboxing):

### Aturan Larangan (Strict Boundaries)
- **No I/O:** AI dilarang import package `os`, `io`, `net/http`, `fmt` (kecuali untuk implementasi log yang diizinkan framework).
- **No Concurrency:** Dilarang menggunakan `go func()`, `sync.Mutex`, atau channel di dalam struct strategi. Strategi harus 100% deterministik dan sinkron (single-threaded state execution).
- **No External State:** Strategi tidak boleh membaca file eksternal atau variabel lingkungan (`os.Getenv`). Segala informasi market harus bersumber dari `sdk.BarContext`.
- **Error Panic Prevention:** Hindari inisialisasi slice atau map tanpa batasan ukuran yang menyebabkan Out-Of-Memory. Lakukan nil/bound check manual jika membuat array custom.

### Aturan SDK
- Harus mengimplementasikan interface `strategy.Strategy` dengan benar.
- Eksekusi order harus selalu melalui fungsi SDK yang disediakan (misal: `ctx.MarketBuy()`). AI dilarang mencoba membuat state posisi secara independen untuk bypass balance check.

## 3. Testing Requirements

- **Target Coverage (Core):** Minimal 80% coverage pada core engine dan SDK pipeline.
- **Auto-Test (AI Strategy):** Setiap file `.go` hasil generasi AI wajib dibuatkan *harness test* (unit test kecil) oleh sistem yang mencekok data dummy kosong (`OHLCV{}`) dan ekstrem, guna memastikan logika `OnBar` AI tidak *panic* (misal karena divide-by-zero).
- **Race Detector:** Gunakan flag `-race` pada testing untuk core runtime concurrent.

```bash
# Standard testing execution
go test -v -race -coverprofile=coverage.out ./...
```

## 4. Linter & Validator Khusus

- Proyek ini akan menggunakan parser AST kustom (via package `go/ast`) untuk menyeleksi dan mereject source code buatan AI jika melanggar batasan import di atas.