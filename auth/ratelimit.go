package auth

import (
	"net/http"
	"sync"
	"time"

	"pxe-manager/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type ipEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type IPRateLimiter struct {
	mu       sync.RWMutex
	ips      map[string]*ipEntry
	requests int
}

func NewIPRateLimiter(requestsPerMinute int) *IPRateLimiter {
	return &IPRateLimiter{ips: make(map[string]*ipEntry), requests: requestsPerMinute}
}

func (i *IPRateLimiter) getLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()
	entry, ok := i.ips[ip]
	if !ok {
		lim := rate.NewLimiter(rate.Every(time.Minute/time.Duration(i.requests)), i.requests)
		entry = &ipEntry{limiter: lim, lastSeen: time.Now()}
		i.ips[ip] = entry
	} else {
		entry.lastSeen = time.Now()
	}
	// 简单 TTL 清理（10分钟未访问移除）
	for k, v := range i.ips {
		if time.Since(v.lastSeen) > 10*time.Minute {
			delete(i.ips, k)
		}
	}
	return entry.limiter
}

func RateLimitMiddleware(sec *SecurityConfig) gin.HandlerFunc {
	rateLimiter := NewIPRateLimiter(sec.RateLimit)
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		// 白名单 IP 不限流
		if utils.IsIPInWhitelist(clientIP, sec.WhitelistIPs) {
			c.Next()
			return
		}
		limiter := rateLimiter.getLimiter(clientIP)
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":  "请求过于频繁，请稍后重试",
				"limit":  sec.RateLimit,
				"window": "每分钟",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
