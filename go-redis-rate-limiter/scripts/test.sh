#!/usr/bin/env bash
set -e

echo "Fixed window:"
for i in {1..7}; do
  curl -s -i \
   -H "X-Client-ID: demo-fixed" \
    -H "X-Rate-Limit-Algorithm: fixed" \
    http://localhost:8080/api | grep -Ei "HTTP/|X-Ratelimit|Retry-After|message|error"
  echo
done

echo "Sliding window:"
for i in {1..7}; do
  curl -s -i \
  -H "X-Client-ID: demo-sliding" \
    -H "X-Rate-Limit-Algorithm: sliding" \
    http://localhost:8080/api | grep -E "HTTP/|X-RateLimit|Retry-After|message|error"
  echo
done
