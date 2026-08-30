package indicators

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"sync"
)

// Cache provides thread-safe caching for indicator calculations
type Cache struct {
	mu    sync.RWMutex
	store map[string]interface{}
}

// NewCache creates a new indicator cache
func NewCache() *Cache {
	return &Cache{
		store: make(map[string]interface{}),
	}
}

// Get retrieves a cached value
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, exists := c.store[key]
	return val, exists
}

// Set stores a value in cache
func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = value
}

// Clear removes all cached values
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store = make(map[string]interface{})
}

// Size returns the number of cached items
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.store)
}

// CachedSMA calculates SMA with caching
func CachedSMA(cache *Cache, data []float64, period int) ([]float64, error) {
	key := generateKey("sma", data, period)
	
	if cached, exists := cache.Get(key); exists {
		return cached.([]float64), nil
	}
	
	result, err := SMA(data, period)
	if err != nil {
		return nil, err
	}
	
	cache.Set(key, result)
	return result, nil
}

// CachedEMA calculates EMA with caching
func CachedEMA(cache *Cache, data []float64, period int) ([]float64, error) {
	key := generateKey("ema", data, period)
	
	if cached, exists := cache.Get(key); exists {
		return cached.([]float64), nil
	}
	
	result, err := EMA(data, period)
	if err != nil {
		return nil, err
	}
	
	cache.Set(key, result)
	return result, nil
}

// CachedRSI calculates RSI with caching
func CachedRSI(cache *Cache, data []float64, period int) ([]float64, error) {
	key := generateKey("rsi", data, period)
	
	if cached, exists := cache.Get(key); exists {
		return cached.([]float64), nil
	}
	
	result, err := RSI(data, period)
	if err != nil {
		return nil, err
	}
	
	cache.Set(key, result)
	return result, nil
}

// CachedMACD calculates MACD with caching
func CachedMACD(cache *Cache, data []float64, fastPeriod, slowPeriod, signalPeriod int) (*MACDResult, error) {
	key := fmt.Sprintf("macd_%x_%d_%d_%d", hashFloats(data), fastPeriod, slowPeriod, signalPeriod)
	
	if cached, exists := cache.Get(key); exists {
		return cached.(*MACDResult), nil
	}
	
	result, err := MACD(data, fastPeriod, slowPeriod, signalPeriod)
	if err != nil {
		return nil, err
	}
	
	cache.Set(key, result)
	return result, nil
}

// CachedATR calculates ATR with caching
func CachedATR(cache *Cache, high, low, close []float64, period int) ([]float64, error) {
	key := fmt.Sprintf("atr_%x_%x_%x_%d", 
		hashFloats(high), hashFloats(low), hashFloats(close), period)
	
	if cached, exists := cache.Get(key); exists {
		return cached.([]float64), nil
	}
	
	result, err := ATR(high, low, close, period)
	if err != nil {
		return nil, err
	}
	
	cache.Set(key, result)
	return result, nil
}

// CachedBollinger calculates Bollinger Bands with caching
func CachedBollinger(cache *Cache, data []float64, period int, stdDev float64) (*BollingerBands, error) {
	key := fmt.Sprintf("bb_%x_%d_%.2f", hashFloats(data), period, stdDev)
	
	if cached, exists := cache.Get(key); exists {
		return cached.(*BollingerBands), nil
	}
	
	result, err := Bollinger(data, period, stdDev)
	if err != nil {
		return nil, err
	}
	
	cache.Set(key, result)
	return result, nil
}

// generateKey creates a cache key from indicator name, data, and period
func generateKey(name string, data []float64, period int) string {
	hash := hashFloats(data)
	return fmt.Sprintf("%s_%x_%d", name, hash, period)
}

// hashFloats creates a hash of float64 slice for cache key
func hashFloats(data []float64) []byte {
	h := sha256.New()
	
	// Write length
	binary.Write(h, binary.LittleEndian, int64(len(data)))
	
	// Sample data points for hash (first, last, middle, and a few random)
	// This is faster than hashing all data for large datasets
	if len(data) > 0 {
		binary.Write(h, binary.LittleEndian, data[0])
		binary.Write(h, binary.LittleEndian, data[len(data)-1])
		
		if len(data) > 2 {
			binary.Write(h, binary.LittleEndian, data[len(data)/2])
		}
		
		// Sample a few more points
		step := len(data) / 10
		if step > 0 {
			for i := step; i < len(data); i += step {
				binary.Write(h, binary.LittleEndian, data[i])
			}
		}
	}
	
	return h.Sum(nil)[:16] // Use first 16 bytes
}

// IndicatorCache is a global cache instance for convenience
var globalCache = NewCache()

// GetGlobalCache returns the global cache instance
func GetGlobalCache() *Cache {
	return globalCache
}

// ClearGlobalCache clears the global cache
func ClearGlobalCache() {
	globalCache.Clear()
}

// CacheStats provides cache statistics
type CacheStats struct {
	Size int
	Hits int
	Misses int
}

// CacheWithStats provides a cache with hit/miss tracking
type CacheWithStats struct {
	cache  *Cache
	hits   int
	misses int
	mu     sync.RWMutex
}

// NewCacheWithStats creates a cache with statistics tracking
func NewCacheWithStats() *CacheWithStats {
	return &CacheWithStats{
		cache: NewCache(),
	}
}

// Get retrieves a value and tracks statistics
func (c *CacheWithStats) Get(key string) (interface{}, bool) {
	val, exists := c.cache.Get(key)
	
	c.mu.Lock()
	if exists {
		c.hits++
	} else {
		c.misses++
	}
	c.mu.Unlock()
	
	return val, exists
}

// Set stores a value in cache
func (c *CacheWithStats) Set(key string, value interface{}) {
	c.cache.Set(key, value)
}

// Stats returns cache statistics
func (c *CacheWithStats) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	return CacheStats{
		Size:   c.cache.Size(),
		Hits:   c.hits,
		Misses: c.misses,
	}
}

// HitRate returns the cache hit rate as a percentage
func (c *CacheWithStats) HitRate() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	total := c.hits + c.misses
	if total == 0 {
		return 0
	}
	
	return float64(c.hits) / float64(total) * 100.0
}

// Clear clears the cache and resets statistics
func (c *CacheWithStats) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.cache.Clear()
	c.hits = 0
	c.misses = 0
}
