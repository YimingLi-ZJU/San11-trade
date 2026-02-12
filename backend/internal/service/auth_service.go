package service

import (
	"errors"
	"time"

	"san11-trade/internal/config"
	"san11-trade/internal/database"
	"san11-trade/internal/model"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrUserExists        = errors.New("username already exists")
	ErrInvalidCredential = errors.New("invalid username or password")
	ErrUserNotFound      = errors.New("user not found")
)

// Claims represents JWT claims
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

// Register creates a new user
func Register(username, password, nickname string) (*model.User, error) {
	db := database.GetDB()

	// Check if username already exists
	var existingUser model.User
	if err := db.Where("username = ?", username).First(&existingUser).Error; err == nil {
		return nil, ErrUserExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username: username,
		Password: string(hashedPassword),
		Nickname: nickname,
		Space:    config.AppConfig.Game.InitialSpace,
	}

	if err := db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates a user and returns JWT token
func Login(username, password string) (string, *model.User, error) {
	db := database.GetDB()

	var user model.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, ErrInvalidCredential
		}
		return "", nil, err
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, ErrInvalidCredential
	}

	// Generate JWT token
	token, err := GenerateToken(&user)
	if err != nil {
		return "", nil, err
	}

	return token, &user, nil
}

// GenerateToken generates a JWT token for a user
func GenerateToken(user *model.User) (string, error) {
	cfg := config.AppConfig

	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		IsAdmin:  user.IsAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.JWT.ExpireHour) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "san11-trade",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWT.Secret))
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString string) (*Claims, error) {
	cfg := config.AppConfig

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GetUserByID retrieves a user by ID
func GetUserByID(id uint) (*model.User, error) {
	db := database.GetDB()

	var user model.User
	if err := db.Preload("Club").First(&user, id).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// GetAllUsers retrieves all registered players
func GetAllUsers() ([]model.User, error) {
	db := database.GetDB()

	var users []model.User
	if err := db.Preload("Club").Where("is_registered = ?", true).Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

// UpdateUserProfile updates user's nickname
func UpdateUserProfile(userID uint, nickname string) error {
	db := database.GetDB()
	return db.Model(&model.User{}).Where("id = ?", userID).Update("nickname", nickname).Error
}

// CreateAdmin creates an admin user if not exists
func CreateAdmin(username, password string) error {
	db := database.GetDB()

	var existingUser model.User
	if err := db.Where("username = ?", username).First(&existingUser).Error; err == nil {
		// User exists, update to admin
		return db.Model(&existingUser).Update("is_admin", true).Error
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := &model.User{
		Username: username,
		Password: string(hashedPassword),
		Nickname: "管理员",
		IsAdmin:  true,
		Space:    0,
	}

	return db.Create(admin).Error
}
