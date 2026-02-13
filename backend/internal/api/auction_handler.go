package api

import (
	"net/http"
	"strconv"

	"san11-trade/internal/service"

	"github.com/gin-gonic/gin"
)

// GetAuctionPool returns all auction generals
func GetAuctionPool(c *gin.Context) {
	generals, err := service.GetAuctionPool()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, generals)
}

// GetAuctionResults returns all auction results
func GetAuctionResults(c *gin.Context) {
	results, err := service.GetAuctionResults()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// GetAuctionStats returns auction statistics
func GetAuctionStats(c *gin.Context) {
	stats, err := service.GetAuctionStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// AssignAuctionRequest is the request body for assigning auction
type AssignAuctionRequest struct {
	GeneralID uint   `json:"general_id" binding:"required"`
	UserID    *uint  `json:"user_id"`
	Price     int    `json:"price"`
	Remark    string `json:"remark"`
}

// AssignAuction handles admin auction assignment
func AssignAuction(c *gin.Context) {
	var req AssignAuctionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	serviceReq := &service.AssignAuctionRequest{
		GeneralID: req.GeneralID,
		UserID:    req.UserID,
		Price:     req.Price,
		Remark:    req.Remark,
	}

	record, err := service.AssignAuction(serviceReq)
	if err != nil {
		status := http.StatusInternalServerError
		if err == service.ErrAuctionGeneralNotFound {
			status = http.StatusNotFound
		} else if err == service.ErrGeneralAlreadyAuctioned {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "拍卖录入成功",
		"record":  record,
	})
}

// ResetAuction resets an auction record
func ResetAuction(c *gin.Context) {
	generalIDStr := c.Param("generalId")
	generalID, err := strconv.ParseUint(generalIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid general id"})
		return
	}

	if err := service.ResetAuctionByGeneralID(uint(generalID)); err != nil {
		status := http.StatusInternalServerError
		if err == service.ErrAuctionRecordNotFound {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "拍卖记录已重置"})
}
