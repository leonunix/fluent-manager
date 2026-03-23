package cache

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/fluent-manager/fluent-manager/internal/config"
	"github.com/redis/go-redis/v9"
)

var (
	client  *redis.Client
	enabled bool
	ttl     time.Duration
)

// Init initialises the Redis client. If the connection fails or caching is
// disabled in config, all cache operations gracefully degrade to no-ops.
func Init(cfg *config.CacheConfig) {
	if !cfg.Enabled {
		log.Println("Cache disabled, metrics will be queried from DB on every request")
		return
	}

	client = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("WARNING: Redis connection failed (%s), falling back to no-cache mode: %v", cfg.Addr, err)
		client = nil
		return
	}

	enabled = true
	ttl = time.Duration(cfg.TTL) * time.Second
	if ttl == 0 {
		ttl = 30 * time.Second
	}
	log.Printf("Redis cache enabled (addr=%s, ttl=%s)", cfg.Addr, ttl)
}

// Get retrieves a cached value. Returns false on miss or when cache is disabled.
func Get(key string, dest interface{}) bool {
	if !enabled {
		return false
	}
	data, err := client.Get(context.Background(), key).Bytes()
	if err != nil {
		return false
	}
	return json.Unmarshal(data, dest) == nil
}

// Set stores a value in cache with the configured TTL.
func Set(key string, value interface{}) {
	if !enabled {
		return
	}
	data, err := json.Marshal(value)
	if err != nil {
		return
	}
	client.Set(context.Background(), key, data, ttl)
}
