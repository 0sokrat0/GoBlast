package middleware

import (
	"GoBlast/pkg/logger"
	"GoBlast/pkg/metrics"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		method := c.Request.Method
		endpoint := c.FullPath()

		c.Next()

		duration := time.Since(start)

		// Запись метрик
		metrics.RequestCounter.WithLabelValues(method, endpoint).Inc()
		metrics.RequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())

		// Запись в лог
		logger.Log.Info("HTTP Request",
			zap.String("method", method),
			zap.String("path", endpoint),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", duration),
			zap.String("client_ip", c.ClientIP()),
		)
	}
}
