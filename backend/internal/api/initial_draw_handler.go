package api

import (
	"net/http"

	"san11-trade/internal/service"

	"github.com/gin-gonic/gin"
)

// InitialDraw handles initial draw request
func InitialDraw(c *gin.Context) {
	userID := GetCurrentUserID(c)

	general, drawType, err := service.InitialDraw(userID)
	if err != nil {
		status := http.StatusBadRequest
		if err == service.ErrNotInInitialDrawPhase {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "初抽成功",
		"general":   general,
		"draw_type": drawType,
	})
}

// GetInitialDrawStatus returns the initial draw status for current user
func GetInitialDrawStatus(c *gin.Context) {
	userID := GetCurrentUserID(c)

	status, err := service.GetInitialDrawStatus(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetAllInitialDrawResults returns all players' initial draw results
func GetAllInitialDrawResults(c *gin.Context) {
	results, err := service.GetAllInitialDrawResults()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// GetInitialDrawPool returns available generals in initial draw pool
func GetInitialDrawPool(c *gin.Context) {
	poolType := c.Query("type")
	if poolType != "initial_guarantee" && poolType != "initial_normal" {
		// Return both pools if type not specified
		guarantee, _ := service.GetInitialDrawPool("initial_guarantee")
		normal, _ := service.GetInitialDrawPool("initial_normal")
		c.JSON(http.StatusOK, gin.H{
			"guarantee": guarantee,
			"normal":    normal,
		})
		return
	}

	generals, err := service.GetInitialDrawPool(poolType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, generals)
}
