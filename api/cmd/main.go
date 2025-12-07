package main

import (
	"log"
	"net/http"

	"p2m-lite/config"
	"p2m-lite/database"
	"p2m-lite/internal/auth"
	"p2m-lite/internal/ws"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load Configuration
	cfg := config.LoadConfig()

	// 2. Initialize Database (Mocked for this example)
	db := database.NewMockDB()

	// 3. Setup Gin Router
	r := gin.Default()
	r.Use(gin.Logger()) // Use default logger for clearer startup

	// 4. Register Modular Routes
	auth.SetupRoutes(r, cfg, db)

	// WebSocket Route
	r.GET("/logs", func(c *gin.Context) {
		ws.HandleLogs(c, cfg)
	})
	r.GET("/logs/:Recorder", func(c *gin.Context) {
		ws.HandleLogs(c, cfg)
	})

	// Health Check
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// 5. Run the server
	log.Println("Starting P2M-Lite Auth Server on http://localhost:8080")
	log.Fatal(r.Run(":8080"))
}
