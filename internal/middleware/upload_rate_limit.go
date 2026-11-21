package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"pclsguserver/internal/config"
)

type rateLimitEntry struct {
	count       int
	windowStart time.Time
}

type UploadRateLimiter struct {
	mu          sync.Mutex
	tokenLimits map[string]*rateLimitEntry
	ipLimits    map[string]*rateLimitEntry
}

func NewUploadRateLimiter() *UploadRateLimiter {
	return &UploadRateLimiter{
		tokenLimits: make(map[string]*rateLimitEntry),
		ipLimits:    make(map[string]*rateLimitEntry),
	}
}

func UploadRateLimitMiddleware(limiter *UploadRateLimiter, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenID := c.GetString("uploadTokenId")
		ip := c.ClientIP()

		window := cfg.Upload.RateLimitWindow
		if window == 0 {
			window = time.Minute
		}
		tokenLimit := cfg.Upload.TokenRateLimit
		if tokenLimit == 0 {
			tokenLimit = 10
		}
		ipLimit := cfg.Upload.IPRateLimit
		if ipLimit == 0 {
			ipLimit = 20
		}

		now := time.Now()

		limiter.mu.Lock()
		defer limiter.mu.Unlock()

		if tokenID != "" {
			entry, exists := limiter.tokenLimits[tokenID]
			if !exists || now.Sub(entry.windowStart) >= window {
				limiter.tokenLimits[tokenID] = &rateLimitEntry{count: 1, windowStart: now}
			} else {
				entry.count++
				if entry.count > tokenLimit {
					fail(c, http.StatusTooManyRequests, "RATE_LIMITED", "上传频率超限，请稍后重试")
					return
				}
			}
		}

		entry, exists := limiter.ipLimits[ip]
		if !exists || now.Sub(entry.windowStart) >= window {
			limiter.ipLimits[ip] = &rateLimitEntry{count: 1, windowStart: now}
		} else {
			entry.count++
			if entry.count > ipLimit {
				fail(c, http.StatusTooManyRequests, "RATE_LIMITED", "上传频率超限，请稍后重试")
				return
			}
		}

		c.Next()
	}
}
