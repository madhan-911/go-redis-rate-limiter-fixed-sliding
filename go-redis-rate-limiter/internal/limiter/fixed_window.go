package limiter

import (
	"context"
	"fmt"
	"time"
)

const fixedWindowScript = `
local count = redis.call("INCR", KEYS[1])

if count == 1 then
    redis.call("PEXPIRE", KEYS[1], ARGV[1])
end

local ttl = redis.call("PTTL", KEYS[1])
return {count, ttl}
`

func (l *RateLimiter) allowFixed(ctx context.Context, clientID string) (Result, error) {
	key := fmt.Sprintf("rate_limit:fixed:%s", clientID)

	values, err := l.redis.Eval(
		ctx,
		fixedWindowScript,
		[]string{key},
		l.window.Milliseconds(),
	).Result()
	if err != nil {
		return Result{}, err
	}

	count := int(values.([]interface{})[0].(int64))
	ttlMs := values.([]interface{})[1].(int64)

	remaining := l.limit - count
	if remaining < 0 {
		remaining = 0
	}

	return Result{
		Allowed:   count <= l.limit,
		Limit:     l.limit,
		Remaining: remaining,
		Reset:     time.Duration(ttlMs) * time.Millisecond,
	}, nil
}
