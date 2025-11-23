package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
    service *AuthService
}

func NewHandler(s *AuthService) *AuthHandler {
    return &AuthHandler{service: s}
}

type InitiateRequest struct {
    PublicKey string `json:"public_key" binding:"required"`
}

func (h *AuthHandler) InitiateAuth(c *gin.Context) {
    var req InitiateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
        return
    }

    randomString, base64Value, err := h.service.InitiateChallenge(req.PublicKey)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate authentication"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "challenge": randomString,
        "key_data":  base64Value,
    })
}

type VerifyRequest struct {
    SignedChallenge string `json:"signed_challenge" binding:"required"`
    KeyData         string `json:"key_data" binding:"required"`
}

func (h *AuthHandler) VerifyAuth(c *gin.Context) {
    var req VerifyRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid verification payload: " + err.Error()})
        return
    }

    sessionToken, err := h.service.CompleteAuthAndIssueToken(req.SignedChallenge, req.KeyData)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed: " + err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message":       "Authentication successful",
        "session_token": sessionToken,
    })
}
