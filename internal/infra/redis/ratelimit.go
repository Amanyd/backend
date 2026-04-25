package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Atomic token-bucket via Lua. Returns 1 if allowed, 0 otherwise.
// KEYS[1] = rate-limit key, ARGV[1] = burst, ARGV[2] = rate, ARGV[3] = now (µs)
var luaTokenBucket = redis.NewScript(`
local key     = KEYS[1]
local burst   = tonumber(ARGV[1])
local rate    = tonumber(ARGV[2])
local now     = tonumber(ARGV[3])

local data = redis.call("HMGET", key, "tokens", "last")
local tokens = tonumber(data[1])
local last   = tonumber(data[2])

if tokens == nil then
  tokens = burst
  last   = now
end

local elapsed = (now - last) / 1000000
local refill  = elapsed * rate
tokens = math.min(burst, tokens + refill)

local allowed = 0
if tokens >= 1 then
  tokens  = tokens - 1
  allowed = 1
end

redis.call("HMSET", key, "tokens", tokens, "last", now)
redis.call("EXPIRE", key, math.ceil(burst / rate) + 1)

return allowed
`)

type RateLimiter struct {
	rdb *redis.Client
}

func NewRateLimiter(rdb *redis.Client) *RateLimiter {
	return &RateLimiter{rdb: rdb}
}

func (rl *RateLimiter) AllowRequest(ctx context.Context, key string, rate, burst int) (bool, error) {
	nowMicro := redisNowMicro(ctx, rl.rdb)

	result, err := luaTokenBucket.Run(ctx, rl.rdb, []string{key}, burst, rate, nowMicro).Int()
	if err != nil {
		return false, fmt.Errorf("ratelimit lua: %w", err)
	}
	return result == 1, nil
}

func redisNowMicro(ctx context.Context, rdb *redis.Client) int64 {
	t, err := rdb.Time(ctx).Result()
	if err != nil {
		return 0
	}
	return t.UnixMicro()
}
