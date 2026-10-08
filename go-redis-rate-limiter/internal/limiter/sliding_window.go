package limiter

import (
	"context"
	"crypto/rand"
	"fmt"
	"strconv"
	"time"
)

const slidingWindowScript = `
local key = KEYS[1]
local window = tonumber(ARGV[1])
local limit = tonumber(ARGV[2])
local member = ARGV[3]

local cutoff = now - window

-- Remove requests outside the rolling window.
redis.call("ZREMRANGEBYSCORE", key, 0, cutoff)

-- Count requests currently inside the window.
local count = redis.call("ZCARD", key)

if count >= limit then
    local oldest = redis.call("ZRANGE", key, 0, 0, "WITHSCORES")
    if #oldest > 0 then
        local retryAfter = window - (now - tonumber(oldest[2]))
        return {0, count, retryAfter}
    end
    return {0, count, window}
end

-- Add this request timestamp.
redis.call("ZADD", key, now, member)
redis.call("PEXPIRE", key, window)

return {1, count + 1, 0}
`

func (l *RateLimiter) allowSliding(ctx context.Context, clientID string) (Result, error) {
	key := fmt.Sprintf("rate_limit:sliding:%s", clientID)
	now := time.Now().UnixMilli()
	member := strconv.FormatInt(now, 10) + ":" + randomID()

	values, err := l.redis.Eval(
		ctx,
		slidingWindowScript,
		[]string{key},
		now,
		l.window.Milliseconds(),
		l.limit,
		member,
	).Result()
	if err != nil {
		return Result{}, err
	}

	raw := values.([]interface{})
	allowed := raw[0].(int64) == 1
	count := int(raw[1].(int64))
	retryMs := raw[2].(int64)

	remaining := l.limit - count
	if remaining < 0 {
		remaining = 0
	}

	return Result{
		Allowed:   allowed,
		Limit:     l.limit,
		Remaining: remaining,
		Reset:     time.Duration(retryMs) * time.Millisecond,
	}, nil
}


func randomID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	return fmt.Sprintf("%x", b)
}
