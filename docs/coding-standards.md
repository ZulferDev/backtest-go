# Coding Standards & Testing Guidelines

## 1. Go Style Guide

- Ikuti standard `gofmt` dan `goimports`.
- Gunakan `golangci-lint` sebagai static analysis checker.
- Hindari penggunaan global state atau package-level mutable variables.
- Gunakan `context.Context` untuk timeout/cancellation pada I/O (fetching data).

## 2. Error Handling

- Gunakan errors wrap jika membutuhkan konteks tambahan: `fmt.Errorf("context: %w", err)`.
- Sentinel error harus diexpose di tingkat package jika client perlu membandingkan tipe error.
- Hindari `panic` pada execution path. Gunakan panic hanya pada inisialisasi kritis yang fatal.
- Nil value check wajib dilakukan sebelum mengakses pointer dereference.

## 3. Testing Requirements

- **Target Coverage:** Minimal 80% coverage pada core engine dan pipeline logic.
- **Table-Driven Tests:** Gunakan pattern ini untuk fungsi utilitas dan indikator matematika.
- **Mocking:** Interface-driven mocking untuk testing engine tanpa menyentuh network exchange.
- **Race Detector:** Gunakan flag `-race` pada testing untuk mendeteksi race conditions di runtime concurrent.

```bash
# Standard testing execution
go test -v -race -coverprofile=coverage.out ./...
```

## 4. Documentation

- Setiap fungsi publik wajib memiliki deskripsi singkat.
- Parameter masukan dan nilai kembalian harus didokumentasikan di header fungsi.
- Contoh implementasi strategi diletakkan di subfolder `/strategies` dengan README terpisah.