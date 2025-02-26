package util

import "time"

func TTLExpired(ts, ttl int64) bool {
	if time.Now().Unix()-ts > ttl {
		return true
	}
	return false
}
