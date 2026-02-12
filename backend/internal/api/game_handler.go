package api

import (
	"net/http"
	"strconv"

	"san11-trade/internal/service"

	"github.com/gin-gonic/gin"
)

// GetGamePhase returns current game phase
func GetGamePhase(c *gin.Context) {
	phase, err := service.GetGamePhase()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, phase)
}

// SetGamePhaseRequest represents phase change request
type SetGamePhaseRequest struct {
	Phase       string `json:"phase" binding:"required"`
	RoundNumber int    `json:"round_number"`
	DraftRound  int    `json:"draft_round"`
}

// SetGamePhase changes the current game phase (admin only)
func SetGamePhase(c *gin.Context) {
	var req SetGamePhaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := service.SetGamePhase(req.Phase, req.RoundNumber, req.DraftRound); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "游戏阶段已更新"})
}

// SignUp handles player registration for the season
func SignUp(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if err := service.SignUp(userID); err != nil {
		status := http.StatusBadRequest
		if err == service.ErrNotInSignupPhase {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "报名成功"})
}

// GetRegisteredPlayers returns all registered players
func GetRegisteredPlayers(c *gin.Context) {
	players, err := service.GetRegisteredPlayers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, players)
}

// GetPlayerRoster returns a player's roster
func GetPlayerRoster(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid player id"})
		return
	}

	roster, err := service.GetUserRoster(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "player not found"})
		return
	}

	c.JSON(http.StatusOK, roster)
}

// GetStatistics returns game statistics
func GetStatistics(c *gin.Context) {
	stats, err := service.GetStatistics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// ResetSeason resets the game for a new season (admin only)
func ResetSeason(c *gin.Context) {
	if err := service.ResetSeason(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "赛季已重置"})
}
