package lua

import "github.com/redis/go-redis/v9"

// RateLimit is a sliding window counter script.
// KEYS[1] = previous window key, KEYS[2] = current window key
// ARGV[1] = weight of previous window (float), ARGV[2] = limit, ARGV[3] = seconds until current window ends
//
// Returns: {0, 0} on allow (after incrementing), {-1, retryAfter} on reject.
var RateLimit = redis.NewScript(`
local prev = tonumber(redis.call("GET", KEYS[1]) or "0")
local curr = tonumber(redis.call("GET", KEYS[2]) or "0")
local weight = tonumber(ARGV[1])
local limit = tonumber(ARGV[2])
local ttl_secs = tonumber(ARGV[3])

local rate = prev * weight + curr
if rate >= limit then
	return {-1, ttl_secs}
end

redis.call("INCR", KEYS[2])
redis.call("EXPIRE", KEYS[2], 120)
return {0, 0}
`)
