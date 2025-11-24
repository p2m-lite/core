package auth

import (
	"p2m-lite/config"
	"p2m-lite/database"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, cfg *config.Config, db *database.MockDB) {
    service := NewService(cfg, db)
    handler := NewHandler(service)

    authGroup := r.Group("/auth")
    {
        // /auth/initiate - The client sends its public key to start the flow
        authGroup.POST("/initiate", handler.InitiateAuth)
		authGroup.POST("/verify", handler.VerifyAuth)
    }
}