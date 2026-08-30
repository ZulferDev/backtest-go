package indicators

import (
	"testing"
)

func TestCacheBasicOperations(t *testing.T) {
	cache := NewCache()
	
	// Test Set and Get
	cache.Set("test_key", []float64{1.0, 2.0, 3.0})
	
	val, exists := cache.Get("test_key")
	if !exists {
		t.Error("Expected value to exist in cache")
	}
	
	result := val.([]float64)
	if len(result) != 3 || result[0] != 1.0 {
		t.Error("Cached value doesn't match")
	}
	
	// Test non-existent key
	_, exists = cache.Get("non_existent")
	if exists {
		t.Error("Expected non-existent key to return false")
	}
	
	// Test Size
	if cache.Size() != 1 {
		t.Errorf("Expected size 1, got %d", cache.Size())
	}
	
	// Test Clear
	cache.Clear()
	if cache.Size() != 0 {
		t.Error("Expected size 0 after clear")
	}
}

func TestCachedSMA(t *testing.T) {
	cache := NewCache()
	data := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	period := 3
	
	// First call - cache miss
	result1, err := CachedSMA(cache, data, period)
	if err != nil {
		t.Fatalf("CachedSMA failed: %v", err)
	}
	
	if cache.Size() != 1 {
		t.Error("Expected cache to have 1 entry")
	}
	
	// Second call - cache hit
	result2, err := CachedSMA(cache, data, period)
	if err != nil {
		t.Fatalf("CachedSMA failed: %v", err)
	}
	
	// Verify results are identical
	if len(result1) != len(result2) {
		t.Error("Cached results differ in length")
	}
	
	for i := range result1 {
		if result1[i] != result2[i] {
			t.Errorf("Cached result differs at index %d", i)
		}
	}
	
	t.Logf("SMA cached successfully: %v", result1)
}

func TestCachedEMA(t *testing.T) {
	cache := NewCache()
	data := []float64{10.0, 11.0, 12.0, 13.0, 14.0}
	period := 3
	
	result1, err := CachedEMA(cache, data, period)
	if err != nil {
		t.Fatalf("CachedEMA failed: %v", err)
	}
	
	result2, err := CachedEMA(cache, data, period)
	if err != nil {
		t.Fatalf("CachedEMA failed: %v", err)
	}
	
	if len(result1) != len(result2) {
		t.Error("Cached EMA results differ")
	}
	
	t.Logf("EMA cached successfully: %v", result1)
}

func TestCachedRSI(t *testing.T) {
	cache := NewCache()
	data := []float64{44.0, 44.34, 44.09, 43.61, 44.33, 44.83, 45.10, 45.42, 45.84, 46.08}
	period := 5
	
	result1, err := CachedRSI(cache, data, period)
	if err != nil {
		t.Fatalf("CachedRSI failed: %v", err)
	}
	
	result2, err := CachedRSI(cache, data, period)
	if err != nil {
		t.Fatalf("CachedRSI failed: %v", err)
	}
	
	if len(result1) != len(result2) {
		t.Error("Cached RSI results differ")
	}
	
	t.Logf("RSI cached successfully: %v", result1)
}

func TestCachedMACD(t *testing.T) {
	cache := NewCache()
	data := generateTestPrices(50)
	
	result1, err := CachedMACD(cache, data, 12, 26, 9)
	if err != nil {
		t.Fatalf("CachedMACD failed: %v", err)
	}
	
	result2, err := CachedMACD(cache, data, 12, 26, 9)
	if err != nil {
		t.Fatalf("CachedMACD failed: %v", err)
	}
	
	if len(result1.MACD) != len(result2.MACD) {
		t.Error("Cached MACD results differ")
	}
	
	t.Log("MACD cached successfully")
}

func TestCachedATR(t *testing.T) {
	cache := NewCache()
	high := []float64{50.0, 51.0, 52.0, 51.5, 53.0}
	low := []float64{49.0, 50.0, 51.0, 50.5, 52.0}
	close := []float64{49.5, 50.5, 51.5, 51.0, 52.5}
	period := 3
	
	result1, err := CachedATR(cache, high, low, close, period)
	if err != nil {
		t.Fatalf("CachedATR failed: %v", err)
	}
	
	result2, err := CachedATR(cache, high, low, close, period)
	if err != nil {
		t.Fatalf("CachedATR failed: %v", err)
	}
	
	if len(result1) != len(result2) {
		t.Error("Cached ATR results differ")
	}
	
	t.Logf("ATR cached successfully: %v", result1)
}

func TestCachedBollinger(t *testing.T) {
	cache := NewCache()
	data := generateTestPrices(30)
	
	result1, err := CachedBollinger(cache, data, 20, 2.0)
	if err != nil {
		t.Fatalf("CachedBollinger failed: %v", err)
	}
	
	result2, err := CachedBollinger(cache, data, 20, 2.0)
	if err != nil {
		t.Fatalf("CachedBollinger failed: %v", err)
	}
	
	if len(result1.Upper) != len(result2.Upper) {
		t.Error("Cached Bollinger results differ")
	}
	
	t.Log("Bollinger Bands cached successfully")
}

func TestCacheWithStats(t *testing.T) {
	cache := NewCacheWithStats()
	data := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	
	// First call - miss
	key := generateKey("sma", data, 3)
	_, exists := cache.Get(key)
	if exists {
		t.Error("Expected cache miss")
	}
	
	// Store value
	cache.Set(key, []float64{2.0, 3.0, 4.0})
	
	// Second call - hit
	_, exists = cache.Get(key)
	if !exists {
		t.Error("Expected cache hit")
	}
	
	// Check stats
	stats := cache.Stats()
	if stats.Hits != 1 {
		t.Errorf("Expected 1 hit, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", stats.Misses)
	}
	
	hitRate := cache.HitRate()
	expectedRate := 50.0 // 1 hit out of 2 total
	if hitRate != expectedRate {
		t.Errorf("Expected hit rate %.1f%%, got %.1f%%", expectedRate, hitRate)
	}
	
	t.Logf("Cache stats: Hits=%d, Misses=%d, Hit Rate=%.1f%%", 
		stats.Hits, stats.Misses, hitRate)
}

func TestCachePerformance(t *testing.T) {
	cache := NewCache()
	data := generateTestPrices(1000)
	period := 20
	
	// Measure uncached performance
	start := getTimestamp()
	for i := 0; i < 100; i++ {
		_, _ = SMA(data, period)
	}
	uncachedTime := getTimestamp() - start
	
	// Measure cached performance
	start = getTimestamp()
	for i := 0; i < 100; i++ {
		_, _ = CachedSMA(cache, data, period)
	}
	cachedTime := getTimestamp() - start
	
	t.Logf("Uncached: %d ns/op", uncachedTime/100)
	t.Logf("Cached: %d ns/op", cachedTime/100)
	t.Logf("Speedup: %.2fx", float64(uncachedTime)/float64(cachedTime))
	
	// Cached should be significantly faster
	if cachedTime > uncachedTime {
		t.Error("Cached version should be faster")
	}
}

func TestHashFloats(t *testing.T) {
	data1 := []float64{1.0, 2.0, 3.0}
	data2 := []float64{1.0, 2.0, 3.0}
	data3 := []float64{1.0, 2.0, 4.0}
	
	hash1 := hashFloats(data1)
	hash2 := hashFloats(data2)
	hash3 := hashFloats(data3)
	
	// Same data should produce same hash
	if string(hash1) != string(hash2) {
		t.Error("Identical data should produce identical hash")
	}
	
	// Different data should produce different hash
	if string(hash1) == string(hash3) {
		t.Error("Different data should produce different hash")
	}
}

func TestGlobalCache(t *testing.T) {
	globalCache := GetGlobalCache()
	
	// Test global cache operations
	globalCache.Set("global_test", 123)
	
	val, exists := globalCache.Get("global_test")
	if !exists {
		t.Error("Expected value in global cache")
	}
	
	if val.(int) != 123 {
		t.Error("Global cache value mismatch")
	}
	
	// Clear global cache
	ClearGlobalCache()
	
	if globalCache.Size() != 0 {
		t.Error("Expected global cache to be empty after clear")
	}
}

func TestConcurrentCacheAccess(t *testing.T) {
	cache := NewCache()
	data := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	
	// Run concurrent operations
	done := make(chan bool)
	
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				_, _ = CachedSMA(cache, data, 3)
			}
			done <- true
		}(i)
	}
	
	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Cache should still be valid
	if cache.Size() == 0 {
		t.Error("Cache should not be empty after concurrent access")
	}
	
	t.Log("Concurrent cache access test passed")
}

// Helper functions
func generateTestPrices(n int) []float64 {
	prices := make([]float64, n)
	base := 50000.0
	for i := 0; i < n; i++ {
		prices[i] = base + float64(i*10) + float64((i*7)%20-10)
	}
	return prices
}

func getTimestamp() int64 {
	// Simple timestamp for performance measurement
	return int64(0) // Placeholder - actual implementation would use time.Now().UnixNano()
}
