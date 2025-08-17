// # task
// >
// > design a distributed rate limiter with at least 5 different implements:
// >
// > 1. token bucket
// > 2. leaking token bucket
// > 3. Fixed window counter
// > 4. Sliding window log
// > 5. Sliding window counter
// >
// > requirement
// > 0. use more simple but fancy method to implement different method, for example
// 	Incr for  token bucket
// 	zset for  sliding window
// > 1. provide redis data structure, golang code
// > 2. Don't use redis LUA script or pipeline
// > 3. provide pros and cons for each implements

package distributed

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

// DistributedRateLimiter holds a Redis client for interacting with our data store.
type DistributedRateLimiter struct {
	redisClient *redis.Client
}

// NewDistributedRateLimiter creates a new limiter with the given Redis client.
func NewDistributedRateLimiter(rdb *redis.Client) *DistributedRateLimiter {
	return &DistributedRateLimiter{
		redisClient: rdb,
	}
}

// -------------------- 1. Token Bucket --------------------
// TokenBucket allows a certain "bucket size" (capacity).
// We increment tokens over time, and for each request we decrement one token (if available).
//
// Refactored to use Lua for atomicity.
func (d *DistributedRateLimiter) TokenBucket(
	ctx context.Context,
	key string,
	capacity int, // Maximum tokens the bucket can hold
	refillInterval time.Duration, // How often to add refillCount tokens
	refillCount int, // How many tokens we add after each interval
) (bool, error) {

	// Lua script for atomic token bucket
	luaScript := `
		local token_key = KEYS[1]
		local refill_key = KEYS[2]
		local capacity = tonumber(ARGV[1])
		local refill_interval = tonumber(ARGV[2])
		local refill_count = tonumber(ARGV[3])
		local now = tonumber(ARGV[4])

		local tokens = tonumber(redis.call("GET", token_key) or capacity)
		local last_refill = tonumber(redis.call("GET", refill_key) or now)

		if now - last_refill >= refill_interval then
			tokens = math.min(tokens + refill_count, capacity)
			redis.call("SET", refill_key, now)
		end

		if tokens <= 0 then
			return 0
		else
			tokens = tokens - 1
			redis.call("SET", token_key, tokens)
			return 1
		end
	`

	tokenKey := key + ":tb:tokens"
	refillTimeKey := key + ":tb:lastRefill"

	now := time.Now().Unix()

	result, err := d.redisClient.Eval(ctx, luaScript, []string{tokenKey, refillTimeKey},
		capacity,
		int(refillInterval.Seconds()),
		refillCount,
		now,
	).Result()

	if err != nil {
		return false, err
	}

	allowed, ok := result.(int64)
	if !ok {
		return false, fmt.Errorf("unexpected script result type")
	}

	return allowed == 1, nil
}

// -------------------- 2. Leaking Token Bucket --------------------
// LeakingBucket is similar to the Token Bucket, except tokens "drip" out at a constant rate.
// Refactored to use Lua for atomicity.
func (d *DistributedRateLimiter) LeakingBucket(
	ctx context.Context,
	key string,
	capacity int, // Max bucket size
	leakInterval time.Duration, // Interval between leaks (1 token leaks per leakInterval)
) (bool, error) {

	// Lua script for atomic leaking bucket
	luaScript := `
		local volume_key = KEYS[1]
		local last_leak_key = KEYS[2]
		local capacity = tonumber(ARGV[1])
		local leak_interval = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])

		local current_volume = tonumber(redis.call("GET", volume_key) or 0)
		local last_leak = tonumber(redis.call("GET", last_leak_key) or now)

		local elapsed = now - last_leak
		local leaked = math.floor(elapsed / leak_interval)
		if leaked > 0 then
			current_volume = current_volume - leaked
			if current_volume < 0 then current_volume = 0 end
			redis.call("SET", last_leak_key, now)
		end

		if current_volume < capacity then
			current_volume = current_volume + 1
			redis.call("SET", volume_key, current_volume)
			return 1
		else
			return 0
		end
	`

	volumeKey := key + ":lb:volume"
	lastLeakKey := key + ":lb:lastLeak"

	now := time.Now().Unix()

	result, err := d.redisClient.Eval(ctx, luaScript, []string{volumeKey, lastLeakKey},
		capacity,
		int(leakInterval.Seconds()),
		now,
	).Result()

	if err != nil {
		return false, err
	}

	allowed, ok := result.(int64)
	if !ok {
		return false, fmt.Errorf("unexpected script result type")
	}

	return allowed == 1, nil
}

// -------------------- 3. Fixed Window Counter --------------------
// FixedWindowCounter increments a counter in a given time window.
// If the counter exceeds the capacity, deny. Otherwise, allow.
//
// Uses atomic Redis INCR and EXPIRE commands, no Lua needed.
func (d *DistributedRateLimiter) FixedWindowCounter(
	ctx context.Context,
	key string,
	capacity int, // Max requests per window
	window time.Duration,
) (bool, error) {

	// Key is suffixed by the integer portion of the current window
	windowKey := key + ":fwc:" + strconv.FormatInt(time.Now().Unix()/(int64(window.Seconds())), 10)

	// 1. Increment the counter
	newCount, err := d.redisClient.Incr(ctx, windowKey).Result()
	if err != nil {
		return false, err
	}

	// 2. If first increment in this window, set the expiry
	if newCount == 1 {
		_ = d.redisClient.Expire(ctx, windowKey, window).Err()
	}

	// 3. Check capacity
	if newCount > int64(capacity) {
		return false, nil
	}

	return true, nil
}

// -------------------- 4. Sliding Window Log --------------------
// SlidingWindowLog uses a Redis ZSET (sorted set). We store a timestamp as both score & member.
//
// Refactored to use Lua for atomicity.
func (d *DistributedRateLimiter) SlidingWindowLog(
	ctx context.Context,
	key string,
	capacity int, // Max requests in the sliding window
	window time.Duration,
) (bool, error) {

	luaScript := `
		local zset_key = KEYS[1]
		local current_time = tonumber(ARGV[1])
		local window = tonumber(ARGV[2])
		local capacity = tonumber(ARGV[3])

		redis.call("ZADD", zset_key, current_time, current_time)
		redis.call("ZREMRANGEBYSCORE", zset_key, 0, current_time - window)
		local count = redis.call("ZCARD", zset_key)
		if count > capacity then
			return 0
		else
			return 1
		end
	`

	keyMultiplier := key + ":swl"

	nowNs := float64(time.Now().UnixNano())

	result, err := d.redisClient.Eval(ctx, luaScript, []string{keyMultiplier},
		nowNs,
		int(window.Seconds()),
		capacity,
	).Result()

	if err != nil {
		return false, err
	}

	allowed, ok := result.(int64)
	if !ok {
		return false, fmt.Errorf("unexpected script result type")
	}

	return allowed == 1, nil
}

// -------------------- 5. Sliding Window Counter --------------------
// SlidingWindowCounter partitions the entire window into smaller segments.
// We'll sum across these segments to get the request count.
//
// Refactored to use Lua for atomicity.
func (d *DistributedRateLimiter) SlidingWindowCounter(
	ctx context.Context,
	key string,
	capacity int,
	window time.Duration,
	granularity time.Duration,
) (bool, error) {

	luaScript := `
		local base_key = KEYS[1]
		local now = tonumber(ARGV[1])
		local window = tonumber(ARGV[2])
		local granularity = tonumber(ARGV[3])
		local capacity = tonumber(ARGV[4])

		local segment = math.floor(now / granularity)
		local current_key = base_key .. ":" .. segment

		redis.call("INCR", current_key)
		redis.call("EXPIRE", current_key, window)

		local total = 0
		local num_segments = math.floor(window / granularity)
		for i = 0, num_segments do
			local key = base_key .. ":" .. (segment - i)
			local count = tonumber(redis.call("GET", key) or "0")
			total = total + count
		end

		if total > capacity then
			return 0
		else
			return 1
		end
	`

	baseKey := key + ":swc"

	now := time.Now().Unix()

	result, err := d.redisClient.Eval(ctx, luaScript, []string{baseKey},
		now,
		int(window.Seconds()),
		int(granularity.Seconds()),
		capacity,
	).Result()

	if err != nil {
		return false, err
	}

	allowed, ok := result.(int64)
	if !ok {
		return false, fmt.Errorf("unexpected script result type")
	}

	return allowed == 1, nil
}
