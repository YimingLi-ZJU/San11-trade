package api

import (
	"net/http"
	"strconv"

	"san11-trade/internal/service"

	"github.com/gin-gonic/gin"
)

// CreateTrade handles trade creation
func CreateTrade(c *gin.Context) {
	userID := GetCurrentUserID(c)
	var req service.TradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trade, err := service.CreateTrade(userID, &req)
	if err != nil {
		status := http.StatusBadRequest
		if err == service.ErrNotInTradingPhase {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "交易请求已发送",
		"trade":   trade,
	})
}

// GetPendingTrades returns pending trades for current user
func GetPendingTrades(c *gin.Context) {
	userID := GetCurrentUserID(c)
	trades, err := service.GetPendingTrades(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, trades)
}

// GetTradeHistory returns trade history for current user
func GetTradeHistory(c *gin.Context) {
	userID := GetCurrentUserID(c)
	trades, err := service.GetTradeHistory(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, trades)
}

// GetAllTrades returns all trades (admin only)
func GetAllTrades(c *gin.Context) {
	trades, err := service.GetAllTrades()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, trades)
}

// GetTradeByID returns a trade by ID
func GetTradeByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid trade id"})
		return
	}

	trade, err := service.GetTradeByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "trade not found"})
		return
	}

	c.JSON(http.StatusOK, trade)
}

// AcceptTrade handles trade acceptance
func AcceptTrade(c *gin.Context) {
	userID := GetCurrentUserID(c)
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid trade id"})
		return
	}

	if err := service.AcceptTrade(uint(id), userID); err != nil {
		status := http.StatusBadRequest
		if err == service.ErrNotTradeParticipant {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "交易已接受"})
}

// RejectTrade handles trade rejection
func RejectTrade(c *gin.Context) {
	userID := GetCurrentUserID(c)
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid trade id"})
		return
	}

	if err := service.RejectTrade(uint(id), userID); err != nil {
		status := http.StatusBadRequest
		if err == service.ErrNotTradeParticipant {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "交易已拒绝"})
}

// CancelTrade handles trade cancellation
func CancelTrade(c *gin.Context) {
	userID := GetCurrentUserID(c)
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid trade id"})
		return
	}

	if err := service.CancelTrade(uint(id), userID); err != nil {
		status := http.StatusBadRequest
		if err == service.ErrNotTradeParticipant {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "交易已取消"})
}
