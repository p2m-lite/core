package api

import (
	"net/http"
	"strconv"
	"time"

	"p2m-lite/internal/database"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.GET("/logs/history", GetLogsHistory)
		api.GET("/recorders", GetRecorders)
	}
}

func GetLogsHistory(c *gin.Context) {
	recorder := c.Query("recorder")
	daysStr := c.Query("days")

	if recorder == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "recorder is required"})
		return
	}

	days := 7 // Default
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil {
			days = d
		}
	}

	cutoff := time.Now().AddDate(0, 0, -days).Unix()

	type DataPoint struct {
		PH        int   `json:"ph"`
		Turbidity int   `json:"turbidity"`
		Timestamp int64 `json:"timestamp"`
	}

	var logs []database.Log
	if err := database.DB.Where("recorder = ? AND timestamp > ?", recorder, cutoff).Order("timestamp asc").Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch logs"})
		return
	}

	var history []DataPoint
	for _, log := range logs {
		history = append(history, DataPoint{
			PH:        log.Ph,
			Turbidity: log.Turbidity,
			Timestamp: log.Timestamp,
		})
	}

	c.JSON(http.StatusOK, gin.H{"history": history})
}

func GetRecorders(c *gin.Context) {
	type RecorderResponse struct {
		Address string  `json:"address"`
		Lat     float64 `json:"lat"`
		Lon     float64 `json:"lon"`
	}

	var recorders []database.Recorder
	if err := database.DB.Find(&recorders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recorders"})
		return
	}

	var response []RecorderResponse
	for _, r := range recorders {
		response = append(response, RecorderResponse{
			Address: r.Address,
			Lat:     r.Lat,
			Lon:     r.Lon,
		})
	}

	c.JSON(http.StatusOK, gin.H{"recorders": response})
}
