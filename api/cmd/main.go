package main

import (
	"log"
	"net/http"

	"p2m-lite/config"
	"p2m-lite/internal/api"
	"p2m-lite/internal/auth"
	"p2m-lite/internal/database"
	"p2m-lite/internal/worker"
	"p2m-lite/internal/ws"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load Configuration
	cfg := config.LoadConfig()

	// 2. Initialize Database
	// Initialize SQLite (Real DB)
	database.InitDB("p2m.db")
	store := database.NewStore(database.DB)

	// 3. Start Background Workers
	worker.StartListener(cfg)
	worker.StartAnalyzer(cfg)

	// 4. Setup Gin Router
	r := gin.Default()
	r.Use(gin.Logger()) // Use default logger for clearer startup

	// 5. Register Modular Routes
	auth.SetupRoutes(r, cfg, store)
	api.SetupRoutes(r)

	// 6. WebSocket Route
	r.GET("/logs", func(c *gin.Context) {
		ws.HandleLogs(c, cfg)
	})
	r.GET("/logs/:Recorder", func(c *gin.Context) {
		ws.HandleLogs(c, cfg)
	})

	// 6. Health Check
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// 7. Run the server
	log.Println("Starting P2M-Lite Auth Server on http://localhost:8080")
	log.Fatal(r.Run(":8080"))
}
