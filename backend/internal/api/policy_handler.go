package api

import (
	"net/http"
	"strconv"
	"time"

	"san11-trade/internal/service"

	"github.com/gin-gonic/gin"
)

// ===== Player APIs =====

// GetPolicyStatus returns the current policy selection status for a player
func GetPolicyStatus(c *gin.Context) {
	status, err := service.GetPolicySelectionStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, status)
}

// GetMyPolicyBid returns the current user's bid
func GetMyPolicyBid(c *gin.Context) {
	userID := c.GetUint("user_id")

	bid, err := service.GetUserPolicyBid(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	prefs, err := service.GetUserPolicyPreferences(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bid":         bid,
		"preferences": prefs,
	})
}

// PlacePolicyBid places or updates a bid
func PlacePolicyBid(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		BidAmount int `json:"bid_amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := service.PlacePolicyBid(userID, req.BidAmount); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "出价成功"})
}

// SetPolicyPreferences sets the user's preferred club order
func SetPolicyPreferences(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		ClubIDs []uint `json:"club_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := service.SetPolicyPreferences(userID, req.ClubIDs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "偏好设置成功"})
}

// SelectPolicyClub selects a club during the selection phase
func SelectPolicyClub(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		ClubID uint `json:"club_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := service.SelectClub(userID, req.ClubID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "选择成功"})
}

// GetPolicySelectionResults returns all selection results
func GetPolicySelectionResults(c *gin.Context) {
	selections, err := service.GetAllPolicySelections()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, selections)
}

// GetClubsWithFilters returns clubs with optional filters
func GetClubsWithFilters(c *gin.Context) {
	league := c.Query("league")
	tag := c.Query("tag")

	var clubs []interface{}
	var err error

	if tag != "" {
		result, e := service.GetClubsByTag(tag)
		err = e
		for _, club := range result {
			clubs = append(clubs, club)
		}
	} else if league != "" {
		result, e := service.GetClubsByLeague(league)
		err = e
		for _, club := range result {
			clubs = append(clubs, club)
		}
	} else {
		result, e := service.GetClubsWithTags()
		err = e
		for _, club := range result {
			clubs = append(clubs, club)
		}
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, clubs)
}

// GetClubFilters returns available leagues and tags
func GetClubFilters(c *gin.Context) {
	leagues, err := service.GetAllLeagues()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tags, err := service.GetAllTags()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"leagues": leagues,
		"tags":    tags,
	})
}

// ===== Admin APIs =====

// AdminClosePolicyBidding closes the bidding phase
func AdminClosePolicyBidding(c *gin.Context) {
	if err := service.CloseBidding(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "出价已截止，已计算选择顺序"})
}

// AdminStartPolicySelection starts the selection phase
func AdminStartPolicySelection(c *gin.Context) {
	var req struct {
		StartTime      string `json:"start_time"` // ISO8601 format
		TimeoutMinutes int    `json:"timeout_minutes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var startTime time.Time
	var err error
	if req.StartTime != "" {
		startTime, err = time.Parse(time.RFC3339, req.StartTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_time format, use ISO8601"})
			return
		}
	} else {
		startTime = time.Now()
	}

	if err := service.StartPolicySelection(startTime, req.TimeoutMinutes); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "选择阶段已开始"})
}

// AdminGetPolicyBids returns all bids (admin only)
func AdminGetPolicyBids(c *gin.Context) {
	bids, err := service.GetAllPolicyBids()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, bids)
}

// AdminResetPolicyPhase resets the entire policy phase
func AdminResetPolicyPhase(c *gin.Context) {
	if err := service.ResetPolicyPhase(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "国策选择已重置"})
}

// AdminResetUserPolicySelection resets a specific user's selection
func AdminResetUserPolicySelection(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	if err := service.ResetUserPolicySelection(uint(userID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户国策选择已重置"})
}

// AdminSelectClubForUser allows admin to select a club for a user
func AdminSelectClubForUser(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var req struct {
		ClubID uint `json:"club_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := service.SelectClub(uint(userID), req.ClubID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已为用户选择国策"})
}

// AdminCheckPolicyTimeout checks and handles timeout (can be called by scheduler)
func AdminCheckPolicyTimeout(c *gin.Context) {
	handled, err := service.CheckAndHandleTimeout()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"timeout_handled": handled,
	})
}

// AdminForceNextSelector forces moving to the next selector (skipping current)
func AdminForceNextSelector(c *gin.Context) {
	config, err := service.GetPolicyPhaseConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if config.CurrentSelector == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no current selector"})
		return
	}

	// Auto-assign for current selector
	if err := service.AutoAssignClub(*config.CurrentSelector); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已跳过当前选择者并自动分配"})
}
