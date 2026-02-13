package api

import (
	"net/http"
	"strings"

	"san11-trade/internal/config"
	"san11-trade/internal/service"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		claims, err := service.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("is_admin", claims.IsAdmin)

		c.Next()
	}
}

// AdminMiddleware checks if user is admin
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get("is_admin")
		if !exists || !isAdmin.(bool) {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RegisteredMiddleware checks if user is registered for the season
func RegisteredMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		user, err := service.GetUserByID(userID.(uint))
		if err != nil || !user.IsRegistered {
			c.JSON(http.StatusForbidden, gin.H{"error": "you must be registered to perform this action"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// GetCurrentUserID extracts user ID from context
func GetCurrentUserID(c *gin.Context) uint {
	userID, _ := c.Get("user_id")
	return userID.(uint)
}

// SetupRouter configures all API routes
func SetupRouter() *gin.Engine {
	if config.AppConfig.Server.Port == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API routes
	api := r.Group("/api")
	{
		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", Register)
			auth.POST("/login", Login)
		}

		// Public routes
		api.GET("/phase", GetGamePhase)
		api.GET("/generals", GetAllGenerals)
		api.GET("/generals/:id", GetGeneralByID)
		api.GET("/treasures", GetAllTreasures)
		api.GET("/treasures/:id", GetTreasureByID)
		api.GET("/clubs", GetAllClubs)
		api.GET("/clubs/:id", GetClubByID)
		api.GET("/clubs/:id/detail", GetClubDetail) // Club with policies
		api.GET("/cities", GetCities)               // City list
		api.GET("/rules", GetGameRules)             // Game rules
		api.GET("/players", GetRegisteredPlayers)
		api.GET("/players/:id/roster", GetPlayerRoster)
		api.GET("/statistics", GetStatistics)

		// Protected routes (require authentication)
		protected := api.Group("")
		protected.Use(AuthMiddleware())
		{
			// User routes
			protected.GET("/me", GetCurrentUser)
			protected.PUT("/me", UpdateProfile)
			protected.GET("/me/roster", GetMyRoster)
			protected.GET("/me/draws", GetMyDrawRecords)
			protected.GET("/me/drafts", GetMyDraftRecords)

			// Game routes (require registration)
			game := protected.Group("")
			game.Use(RegisteredMiddleware())
			{
				// Draw routes
				game.POST("/draw/guarantee", GuaranteeDraw)
				game.POST("/draw/normal", NormalDraw)
				game.GET("/draw/pool/:type", GetDrawPool)

				// Draft routes
				game.GET("/draft/pool", GetDraftPool)
				game.POST("/draft/pick", DraftPick)

				// Trade routes
				game.POST("/trades", CreateTrade)
				game.GET("/trades/pending", GetPendingTrades)
				game.GET("/trades/history", GetTradeHistory)
				game.GET("/trades/:id", GetTradeByID)
				game.POST("/trades/:id/accept", AcceptTrade)
				game.POST("/trades/:id/reject", RejectTrade)
				game.POST("/trades/:id/cancel", CancelTrade)
			}

			// Sign up (doesn't require previous registration)
			protected.POST("/signup", SignUp)
		}

		// Admin routes
		admin := api.Group("/admin")
		admin.Use(AuthMiddleware(), AdminMiddleware())
		{
			admin.POST("/phase", SetGamePhase)
			admin.POST("/reset", ResetSeason)
			admin.GET("/trades", GetAllTrades)
			admin.POST("/import", ImportData)
		}
	}

	return r
}
