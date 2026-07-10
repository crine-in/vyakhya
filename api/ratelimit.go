// Copyright (C) 2026 CRINE (https://www.crine.in) <support@crine.in>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package api

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ipRateLimiter implements a thread-safe token-bucket rate limiter per IP.
type ipRateLimiter struct {
	mu         sync.Mutex
	ips        map[string]*ipClient
	rate       float64 // tokens refilled per second
	burst      float64 // max tokens allowed at once
	lastEvict  time.Time
}

type ipClient struct {
	tokens   float64
	lastSeen time.Time
}

// newIPRateLimiter creates a new rate limiter with the given rps (refill rate) and burst size.
func newIPRateLimiter(rate float64, burst int) *ipRateLimiter {
	return &ipRateLimiter{
		ips:       make(map[string]*ipClient),
		rate:      rate,
		burst:     float64(burst),
		lastEvict: time.Now(),
	}
}

// allow checks if a request from the given IP is allowed under the rate limit.
func (lim *ipRateLimiter) allow(ip string) bool {
	lim.mu.Lock()
	defer lim.mu.Unlock()

	now := time.Now()

	// Periodic eviction of stale clients every 5 minutes to prevent memory leaks
	if now.Sub(lim.lastEvict) > 5*time.Minute {
		for k, v := range lim.ips {
			if now.Sub(v.lastSeen) > 10*time.Minute {
				delete(lim.ips, k)
			}
		}
		lim.lastEvict = now
	}

	client, exists := lim.ips[ip]
	if !exists {
		client = &ipClient{
			tokens:   lim.burst,
			lastSeen: now,
		}
		lim.ips[ip] = client
	}

	// Calculate elapsed time and add refilled tokens
	elapsed := now.Sub(client.lastSeen).Seconds()
	client.lastSeen = now
	client.tokens += elapsed * lim.rate
	if client.tokens > lim.burst {
		client.tokens = lim.burst
	}

	// Consume 1 token if available
	if client.tokens >= 1.0 {
		client.tokens -= 1.0
		return true
	}
	return false
}

// getIP extracts the client IP address from the request headers or remote address.
func getIP(r *http.Request) string {
	// 1. Check X-Forwarded-For (set by proxies/load balancers)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}

	// 2. Check X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		ip := strings.TrimSpace(xri)
		if ip != "" {
			return ip
		}
	}

	// 3. Fallback to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
