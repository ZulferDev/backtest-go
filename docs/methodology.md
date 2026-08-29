# Research Methodology — backtest-go (Code Creator AI)

**Version:** 2.0  
**Status:** Active

---

## Overview

Framework ini memperlakukan AI sebagai **Quantitative Developer & Researcher**. Metode riset ini bukan tentang memutar knob parameter (parameter optimization), tetapi tentang **Discovery of Logic**. AI merumuskan hipotesa pasar dan mengekspresikannya dalam bentuk **Kode Golang** yang mengontrol logika entry/exit/manajemen risiko.

---

## The Research Cycle (AI-Driven)

### Phase 1: Observation & Hypothesis Formation

**AI Action:**
AI menggunakan internet search, data historis statis, atau literature (jurnal kuantitatif) untuk menemukan anomali pasar.

**Output:** 
Hipotesa testable dalam bentuk narasi (Markdown).

*Contoh:* 
> "Karena sifat mean-reverting crypto di akhir pekan, saya berhipotesa bahwa volatilitas tinggi yang dipadu dengan volume menurun pada Jumat malam memicu koreksi. Saya akan membuat custom momentum score yang memperhitungkan volume profile."

---

### Phase 2: Strategy Implementation (Code Generation)

**AI Action:**
AI mengimplementasikan hipotesa ke dalam kode murni Golang, memanfaatkan **Strategy SDK**.
AI tidak membuat JSON; AI membuat logika (if/else, perulangan, perhitungan matematika, pengelolaan array state internal).

**Guidelines untuk AI:**
- Kode harus robust, tangani data nol (`if close == 0`).
- Buat struct internal untuk menyimpan state historis custom jika SDK tidak menyediakan.
- Implementasikan logic risk management secara native di dalam fungsi `OnBar` (contoh: trailing stop custom).

**Output:** 
File `strategies/ai_layer/strategy_<id>.go`

---

### Phase 3: Sandboxing & Compilation Check

**Engine Action:**
Sistem tidak langsung menjalankan backtest. Backtest-go pipeline akan:
1. Mem-parsing AST (Abstract Syntax Tree) dari file AI.
2. Memblokir kode jika terdeteksi import `os`, manipulasi file, eksekusi OS, atau goroutines.
3. Menjalankan quick unit-test (`go test`) untuk memastikan tidak ada *panic* pada logic AI.

**Failure Loop:** Jika gagal compile/test, Engine melempar raw error log kembali ke AI. AI harus memperbaiki syntax/logic-nya dan mencoba ulang (seperti developer pada umumnya).

---

### Phase 4: Historical Simulation (Backtest)

**Engine Action:**
Kode lolos uji diikat (compiled/run) ke dataset jutaan bar. Core engine murni Golang menghitung fills order, memotong komisi, mensimulasikan slippage, dan mencatat equity curve.

**Output:** 
File `results.json` berisi metrik yang *undisputable* (AI tidak bisa mengubah atau mengestimasi hasil PnL sendiri).

---

### Phase 5: Objective Analysis & Refinement (The Learning Loop)

**AI Action:**
AI membaca `results.json`. 
AI menganalisa titik kegagalan (misalnya: "Max drawdown terjadi karena stop-loss statis saya berulang kali tersentuh di fase ranging").

**The Refinement (Pivot or Iterate):**
- **Iterate:** AI memodifikasi file `.go` sebelumnya, menambahkan filter (misal: menambahkan `ctx.Indicator.ADX()` untuk filter ranging market).
- **Pivot:** AI menghapus logika lama, menulis konsep strategi baru dari nol.

---

### Phase 6: Overfitting Prevention (Walk-Forward)

Untuk menghindari AI membuat kode yang sekadar menghafal (curve-fitting) dataset historical, evaluasi performa akan dipisah:

1. AI mengembangkan kode menggunakan data **In-Sample** (contoh: 2022-2023).
2. Saat AI yakin dengan kodenya, sistem memaksakan tes **Out-of-Sample** (contoh: 2024).
3. AI diinstruksikan untuk menganalisa *degradasi performa* antara dua set tersebut.

---

## Documentation Standards (Output dari AI)

Setiap siklus riset AI wajib menyimpan 1 set artefak di direktori `research_logs/`:
1. `hypothesis.md` (Mengapa strategi ini dibuat)
2. `strategy.go` (Kode nyata yang dijalankan)
3. `result.json` (Hasil akhir backtest objketif)
4. `post_mortem.md` (Analisa AI: mengapa strategi ini untung/rugi, apa pelajaran untuk strategi selanjutnya).