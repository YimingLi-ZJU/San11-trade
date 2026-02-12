package api

import (
	"net/http"
	"strconv"

	"san11-trade/internal/service"

	"github.com/gin-gonic/gin"
)

// GuaranteeDraw handles guarantee draw
func GuaranteeDraw(c *gin.Context) {
	userID := GetCurrentUserID(c)
	general, err := service.GuaranteeDraw(userID)
	if err != nil {
		status := http.StatusBadRequest
		if err == service.ErrNotInDrawPhase {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "抽将成功",
		"general": general,
	})
}

// NormalDraw handles normal draw
func NormalDraw(c *gin.Context) {
	userID := GetCurrentUserID(c)
	general, err := service.NormalDraw(userID)
	if err != nil {
		status := http.StatusBadRequest
		if err == service.ErrNotInDrawPhase {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "抽将成功",
		"general": general,
	})
}

// GetDrawPool returns available generals in a pool
func GetDrawPool(c *gin.Context) {
	poolType := c.Param("type")
	if poolType != "guarantee" && poolType != "normal" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pool type"})
		return
	}

	generals, err := service.GetAvailablePoolGenerals(poolType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, generals)
}

// GetDraftPool returns available generals for draft
func GetDraftPool(c *gin.Context) {
	generals, err := service.GetDraftPool()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, generals)
}

// DraftPickRequest represents a draft pick request
type DraftPickRequest struct {
	GeneralID uint `json:"general_id" binding:"required"`
}

// DraftPick handles draft pick
func DraftPick(c *gin.Context) {
	userID := GetCurrentUserID(c)
	var req DraftPickRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	general, err := service.DraftPick(userID, req.GeneralID)
	if err != nil {
		status := http.StatusBadRequest
		if err == service.ErrNotInDraftPhase || err == service.ErrNotYourTurn {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "选秀成功",
		"general": general,
	})
}

// GetAllGenerals returns all generals
func GetAllGenerals(c *gin.Context) {
	generals, err := service.GetAllGenerals()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, generals)
}

// GetGeneralByID returns a general by ID
func GetGeneralByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid general id"})
		return
	}

	general, err := service.GetGeneralByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "general not found"})
		return
	}

	c.JSON(http.StatusOK, general)
}

// GetAllTreasures returns all treasures
func GetAllTreasures(c *gin.Context) {
	treasures, err := service.GetAllTreasures()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, treasures)
}

// GetTreasureByID returns a treasure by ID
func GetTreasureByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid treasure id"})
		return
	}

	treasure, err := service.GetTreasureByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "treasure not found"})
		return
	}

	c.JSON(http.StatusOK, treasure)
}

// GetAllClubs returns all clubs
func GetAllClubs(c *gin.Context) {
	clubs, err := service.GetAllClubs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, clubs)
}

// GetClubByID returns a club by ID
func GetClubByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid club id"})
		return
	}

	club, err := service.GetClubByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "club not found"})
		return
	}

	c.JSON(http.StatusOK, club)
}
