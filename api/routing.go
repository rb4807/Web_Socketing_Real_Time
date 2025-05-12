package api

import (
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rb4807/Web_Socketing_Real_Time/core"
)

func AddRoutes(router *gin.Engine) {
	router.Use(cors.Default())

	apiV1 := router.Group("/api/v1")
	{
		apiV1.GET("/tickers", GetAllTickers)
		apiV1.GET("/consumers", GetActiveConsumers)

		// Add endpoint to get consumer count for a specific ticker
		apiV1.GET("/consumers/:ticker", func(c *gin.Context) {
			ticker := c.Param("ticker")
			topic := "trades-" + strings.ToLower(ticker)
			count := core.Bus.GetConsumerCount(topic)
			c.JSON(200, gin.H{
				"ticker":         ticker,
				"topic":          topic,
				"consumer_count": count,
			})
		})
	}

	// WebSocket route
	router.GET("/ws/trades/:ticker", func(c *gin.Context) {
		if c.Request.Header.Get("Upgrade") != "websocket" {
			c.JSON(400, gin.H{"error": "WebSocket upgrade required"})
			return
		}

		// Specific WebSocket logic here
		ListenTicker(c)
	})
}
