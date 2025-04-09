package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rb4807/Web_Socketing_Real_Time/api"
	"github.com/rb4807/Web_Socketing_Real_Time/core"
)

func main() {
	// Load configuration
	core.Load()
	
	// Initialize message bus (Kafka replacement)
	core.InitMessageBus()
	
	// Start the trade data simulator
	core.StartSimulation()
	
	// Setup the router
	router := gin.Default()

	// CORS middleware
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AllowMethods = []string{"GET", "POST", "HEAD", "PUT", "DELETE", "PATCH", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Content-Length", "Accept-Language", "Accept-Encoding", "Connection", "Access-Control-Allow-Origin"}
	config.AllowCredentials = true
	router.Use(cors.New(config))

	// Add routes
	api.AddRoutes(router)

	// Start server
	router.Run(":8000")
}