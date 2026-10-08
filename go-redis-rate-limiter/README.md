# Go + Redis Distributed Rate Limiter

A learning project implementing two distributed HTTP rate-limiting algorithms:

1. Fixed Window
2. Sliding Window

Both use Redis as the shared state store, so multiple API instances can enforce the same limit.

## Requirements

- Go 1.22+
- Redis 7+
- curl/Postman

## Run Redis

```bash
docker compose up -d
```

## Run the API

```bash
go mod tidy
go run ./cmd/server
```

The API starts on `http://localhost:8080`.

## Try Fixed Window

Limit: 5 requests per 10 seconds.

```bash
curl -H "X-Rate-Limit-Algorithm: fixed" http://localhost:8080/api
```

After 5 successful requests, you should receive HTTP 429 until the fixed window expires.

## Try Sliding Window

```bash
curl -H "X-Rate-Limit-Algorithm: sliding" http://localhost:8080/api
```

The sliding window continuously considers the previous 10 seconds instead of resetting at a fixed boundary.

## Headers

- `X-Client-ID`: identifies the client. Defaults to the remote IP.
- `X-Rate-Limit-Algorithm`: `fixed` or `sliding`
- `X-RateLimit-Limit`: configured request limit
- `X-RateLimit-Remaining`: remaining requests
- `X-RateLimit-Reset`: seconds until the relevant window resets/expires

## Project structure

```text
cmd/server              application entry point
internal/limiter        rate-limiting algorithms and Redis scripts
internal/httpapi        HTTP middleware/handlers
scripts                 small testing helpers
docker-compose.yml      local Redis
```

## Important learning topics

Read `internal/limiter/fixed_window.go` first, then `sliding_window.go`.

Pay attention to:

- Redis `INCR` + `EXPIRE`
- Redis Sorted Sets (`ZADD`, `ZREMRANGEBYSCORE`, `ZCARD`)
- Lua scripts and atomicity
- HTTP middleware
- concurrency
- distributed state
- race conditions
- 429 responses

## Fixed Window vs Sliding Window

Fixed Window stores a counter for each time bucket:

```text
10:00:00 -------- 10:00:10
      one counter
```

Sliding Window stores individual request timestamps:

```text
10:00:01  10:00:04  10:00:08
          <- rolling 10 seconds ->
```

The sliding approach avoids the boundary problem of fixed windows, at the cost of storing more request information.
