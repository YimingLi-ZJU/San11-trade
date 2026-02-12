package api

import (
	"net/http"

	"san11-trade/internal/service"

	"github.com/gin-gonic/gin"
)

// RegisterRequest represents registration request
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
	Nickname string `json:"nickname"`
}

// LoginRequest represents login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Register handles user registration
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Nickname == "" {
		req.Nickname = req.Username
	}

	user, err := service.Register(req.Username, req.Password, req.Nickname)
	if err != nil {
		if err == service.ErrUserExists {
			c.JSON(http.StatusConflict, gin.H{"error": "用户名已存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Generate token for the new user
	token, err := service.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "注册成功",
		"token":   token,
		"user":    user,
	})
}

// Login handles user login
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, user, err := service.Login(req.Username, req.Password)
	if err != nil {
		if err == service.ErrInvalidCredential {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "登录成功",
		"token":   token,
		"user":    user,
	})
}

// GetCurrentUser returns current user info
func GetCurrentUser(c *gin.Context) {
	userID := GetCurrentUserID(c)
	user, err := service.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateProfileRequest represents profile update request
type UpdateProfileRequest struct {
	Nickname string `json:"nickname" binding:"required,max=50"`
}

// UpdateProfile updates user profile
func UpdateProfile(c *gin.Context) {
	userID := GetCurrentUserID(c)
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := service.UpdateUserProfile(userID, req.Nickname); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// GetMyRoster returns current user's roster
func GetMyRoster(c *gin.Context) {
	userID := GetCurrentUserID(c)
	roster, err := service.GetUserRoster(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, roster)
}

// GetMyDrawRecords returns current user's draw records
func GetMyDrawRecords(c *gin.Context) {
	userID := GetCurrentUserID(c)
	records, err := service.GetUserDrawRecords(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, records)
}

// GetMyDraftRecords returns current user's draft records
func GetMyDraftRecords(c *gin.Context) {
	userID := GetCurrentUserID(c)
	records, err := service.GetUserDraftRecords(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, records)
}
