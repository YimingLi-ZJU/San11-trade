package api

import (
	"net/http"
	"strconv"

	"san11-trade/internal/config"
	"san11-trade/internal/service"

	"github.com/gin-gonic/gin"
)

// GenerateInviteCodesRequest represents the request to generate invite codes
type GenerateInviteCodesRequest struct {
	Count      int    `json:"count"`       // Number of codes to generate (default 1, max 100)
	Type       int    `json:"type"`        // 0=single-use, 1=multi-use
	MaxUses    int    `json:"max_uses"`    // Max uses per code (default 1)
	ExpireDays int    `json:"expire_days"` // Days until expiration (0=never)
	Remark     string `json:"remark"`      // Optional remark
}

// GenerateInviteCodes handles POST /api/admin/invite-codes
func GenerateInviteCodes(c *gin.Context) {
	var req GenerateInviteCodesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminID := GetCurrentUserID(c)

	serviceReq := service.GenerateInviteCodeRequest{
		Count:      req.Count,
		Type:       req.Type,
		MaxUses:    req.MaxUses,
		ExpireDays: req.ExpireDays,
		Remark:     req.Remark,
	}

	codes, err := service.GenerateInviteCodes(serviceReq, adminID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成邀请码失败: " + err.Error()})
		return
	}

	// Extract just the code strings for easy copying
	codeStrings := make([]string, len(codes))
	for i, code := range codes {
		codeStrings[i] = code.Code
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "邀请码生成成功",
		"codes":        codes,
		"code_strings": codeStrings,
	})
}

// GetInviteCodes handles GET /api/admin/invite-codes
func GetInviteCodes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	codes, total, err := service.GetAllInviteCodes(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取邀请码列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"codes":     codes,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// DeleteInviteCode handles DELETE /api/admin/invite-codes/:id
func DeleteInviteCode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的邀请码ID"})
		return
	}

	if err := service.DeleteInviteCode(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除邀请码失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "邀请码已删除"})
}

// GetInviteCodeUsages handles GET /api/admin/invite-codes/:id/usages
func GetInviteCodeUsages(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的邀请码ID"})
		return
	}

	usages, err := service.GetInviteCodeUsages(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取使用记录失败"})
		return
	}

	c.JSON(http.StatusOK, usages)
}

// GetInviteCodeStats handles GET /api/admin/invite-codes/stats
func GetInviteCodeStats(c *gin.Context) {
	stats, err := service.GetInviteCodeStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取统计信息失败"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// ValidateInviteCode handles GET /api/invite-codes/validate?code=xxx
// This is a public endpoint to check if a code is valid
func ValidateInviteCode(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供邀请码", "valid": false})
		return
	}

	inviteCode, err := service.GetInviteCodeByCode(code)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"valid": false, "reason": "邀请码不存在"})
		return
	}

	if !inviteCode.IsValid() {
		reason := "邀请码已失效"
		if inviteCode.UsedCount >= inviteCode.MaxUses {
			reason = "邀请码已被使用"
		}
		c.JSON(http.StatusOK, gin.H{"valid": false, "reason": reason})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":     true,
		"remaining": inviteCode.MaxUses - inviteCode.UsedCount,
	})
}

// GetRegistrationConfig handles GET /api/config/registration
// Returns whether invite code is required for registration
func GetRegistrationConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"require_invite_code": config.AppConfig.Registration.RequireInviteCode,
	})
}
