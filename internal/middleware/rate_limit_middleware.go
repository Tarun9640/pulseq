package middleware

import (
	"context"
	"net/http"

	"github.com/Tarun9640/pulseq/internal/ratelimiter"
	"github.com/gin-gonic/gin"
)

func RateLimitMiddleware(limiter *ratelimiter.APIRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		//client ip
		ip := c.ClientIP()

		allowed, err := limiter.Allow(context.Background(), ip)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Rate limiter error",
			})
			c.Abort()
			return 
		}
		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too Many Requests",
			})
			c.Abort()
			return
		}

		c.Next()

	}
}
