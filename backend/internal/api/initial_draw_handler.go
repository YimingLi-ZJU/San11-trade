package api

import (
	"net/http"
	"strconv"

	"san11-trade/internal/model"
	"san11-trade/internal/service"

	"github.com/gin-gonic/gin"
)

// DrawOnce handles draw request
func DrawOnce(c *gin.Context) {
	userID := GetCurrentUserID(c)

	general, drawType, err := service.Draw(userID)
	if err != nil {
		status := http.StatusBadRequest
		if err == service.ErrNotInDrawPhase {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "抽将成功",
		"general":   general,
		"draw_type": drawType,
	})
}

// GetDrawStatus returns the draw status for current user
func GetDrawStatus(c *gin.Context) {
	userID := GetCurrentUserID(c)

	status, err := service.GetDrawStatus(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetAllDrawResults returns all players' draw results
func GetAllDrawResults(c *gin.Context) {
	results, err := service.GetAllDrawResults()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// GetDrawPoolHandler returns available generals in draw pool
func GetDrawPoolHandler(c *gin.Context) {
	poolType := c.Query("type")
	if poolType != "initial_guarantee" && poolType != "initial_normal" {
		// Return both pools if type not specified
		guarantee, _ := service.GetDrawPool("initial_guarantee")
		normal, _ := service.GetDrawPool("initial_normal")
		c.JSON(http.StatusOK, gin.H{
			"guarantee": guarantee,
			"normal":    normal,
		})
		return
	}

	generals, err := service.GetDrawPool(poolType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, generals)
}

// ===== Admin Draw APIs =====

// AdminResetUserDraw resets a user's draw results
func AdminResetUserDraw(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if err := service.ResetUserDraw(uint(userID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户抽将记录已重置"})
}

// AdminResetAllDraw resets all users' draw results
func AdminResetAllDraw(c *gin.Context) {
	count, err := service.ResetAllUsersDraw()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "所有用户抽将记录已重置",
		"reset_count": count,
	})
}

// AdminDrawForUser performs all draws for a specific user
func AdminDrawForUser(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	generals, err := service.DrawForUser(uint(userID))
	if err != nil && err != service.ErrDrawLimitReached {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "代抽完成",
		"generals": generals,
		"count":    len(generals),
	})
}

// AdminDrawForAll performs all draws for all registered users
func AdminDrawForAll(c *gin.Context) {
	results, err := service.DrawForAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to response format
	type UserDrawResult struct {
		UserID   uint            `json:"user_id"`
		Generals []model.General `json:"generals"`
		Count    int             `json:"count"`
	}

	response := make([]UserDrawResult, 0, len(results))
	totalCount := 0
	for userID, generals := range results {
		response = append(response, UserDrawResult{
			UserID:   userID,
			Generals: generals,
			Count:    len(generals),
		})
		totalCount += len(generals)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "批量抽将完成",
		"results":     response,
		"user_count":  len(results),
		"total_count": totalCount,
	})
}
