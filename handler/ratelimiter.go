package handler

import (
	"errors"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type Client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	mu      sync.Mutex
	clients map[string]*Client
}

func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{clients: make(map[string]*Client)}
	rl.startCleaning()
	return rl
}

func (rl *RateLimiter) Allow(ip string) error {
	rl.mu.Lock()

	if _, found := rl.clients[ip]; !found {
		rl.clients[ip] = &Client{limiter: rate.NewLimiter(2, 4)}
	}

	rl.clients[ip].lastSeen = time.Now()

	if !rl.clients[ip].limiter.Allow() {
		rl.mu.Unlock()
		return errors.New("rate limit exceeded")
	}

	rl.mu.Unlock()
	return nil
}

func (rl *RateLimiter) startCleaning() {
	go func() {
		for {
			time.Sleep(time.Minute)
			rl.mu.Lock()
			for ip, client := range rl.clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(rl.clients, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()
}
