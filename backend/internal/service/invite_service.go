package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"san11-trade/internal/database"
	"san11-trade/internal/model"

	"gorm.io/gorm"
)

var (
	ErrInviteCodeNotFound = errors.New("invite code not found")
	ErrInviteCodeInvalid  = errors.New("invite code is invalid or expired")
	ErrInviteCodeUsed     = errors.New("invite code has been fully used")
	ErrInviteCodeRequired = errors.New("invite code is required")
)

// GenerateInviteCodeRequest represents the request to generate invite codes
type GenerateInviteCodeRequest struct {
	Count      int    `json:"count"`       // Number of codes to generate (default 1)
	Type       int    `json:"type"`        // 0=single-use, 1=multi-use
	MaxUses    int    `json:"max_uses"`    // Max uses per code (default 1)
	ExpireDays int    `json:"expire_days"` // Days until expiration (0=never)
	Remark     string `json:"remark"`      // Optional remark
}

// GenerateInviteCodes creates multiple invite codes
func GenerateInviteCodes(req GenerateInviteCodeRequest, createdBy uint) ([]model.InviteCode, error) {
	db := database.GetDB()

	// Set defaults
	if req.Count <= 0 {
		req.Count = 1
	}
	if req.Count > 100 {
		req.Count = 100 // Limit to 100 codes at once
	}
	if req.MaxUses <= 0 {
		req.MaxUses = 1
	}

	var expiredAt *time.Time
	if req.ExpireDays > 0 {
		t := time.Now().AddDate(0, 0, req.ExpireDays)
		expiredAt = &t
	}

	codes := make([]model.InviteCode, 0, req.Count)
	for i := 0; i < req.Count; i++ {
		code, err := generateUniqueCode()
		if err != nil {
			return nil, err
		}

		inviteCode := model.InviteCode{
			Code:      code,
			Type:      req.Type,
			MaxUses:   req.MaxUses,
			UsedCount: 0,
			ExpiredAt: expiredAt,
			CreatedBy: createdBy,
			Remark:    req.Remark,
		}

		if err := db.Create(&inviteCode).Error; err != nil {
			return nil, err
		}

		codes = append(codes, inviteCode)
	}

	return codes, nil
}

// generateUniqueCode generates a unique 16-character hex code
func generateUniqueCode() (string, error) {
	bytes := make([]byte, 8) // 8 bytes = 16 hex characters
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GetAllInviteCodes retrieves all invite codes with pagination
func GetAllInviteCodes(page, pageSize int) ([]model.InviteCode, int64, error) {
	db := database.GetDB()

	var total int64
	db.Model(&model.InviteCode{}).Count(&total)

	var codes []model.InviteCode
	offset := (page - 1) * pageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&codes).Error; err != nil {
		return nil, 0, err
	}

	return codes, total, nil
}

// GetInviteCodeByCode retrieves an invite code by its code string
func GetInviteCodeByCode(code string) (*model.InviteCode, error) {
	db := database.GetDB()

	var inviteCode model.InviteCode
	if err := db.Where("code = ?", code).First(&inviteCode).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInviteCodeNotFound
		}
		return nil, err
	}

	return &inviteCode, nil
}

// ValidateAndUseInviteCode validates an invite code and marks it as used
func ValidateAndUseInviteCode(code string, userID uint) error {
	db := database.GetDB()

	return db.Transaction(func(tx *gorm.DB) error {
		var inviteCode model.InviteCode
		// Lock the row for update
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("code = ?", code).First(&inviteCode).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrInviteCodeNotFound
			}
			return err
		}

		// Check validity
		if !inviteCode.IsValid() {
			return ErrInviteCodeInvalid
		}

		// Update usage count
		inviteCode.UsedCount++
		if err := tx.Save(&inviteCode).Error; err != nil {
			return err
		}

		// Record usage
		usage := model.InviteCodeUsage{
			InviteCodeID: inviteCode.ID,
			UserID:       userID,
			UsedAt:       time.Now(),
		}
		if err := tx.Create(&usage).Error; err != nil {
			return err
		}

		return nil
	})
}

// DeleteInviteCode deletes an invite code by ID
func DeleteInviteCode(id uint) error {
	db := database.GetDB()

	// Delete usage records first
	if err := db.Where("invite_code_id = ?", id).Delete(&model.InviteCodeUsage{}).Error; err != nil {
		return err
	}

	// Delete the invite code
	return db.Delete(&model.InviteCode{}, id).Error
}

// GetInviteCodeUsages retrieves usage records for an invite code
func GetInviteCodeUsages(inviteCodeID uint) ([]model.InviteCodeUsage, error) {
	db := database.GetDB()

	var usages []model.InviteCodeUsage
	if err := db.Preload("User").Where("invite_code_id = ?", inviteCodeID).
		Order("used_at DESC").Find(&usages).Error; err != nil {
		return nil, err
	}

	return usages, nil
}

// GetInviteCodeStats returns statistics about invite codes
func GetInviteCodeStats() (map[string]int64, error) {
	db := database.GetDB()

	stats := make(map[string]int64)

	// Total codes
	db.Model(&model.InviteCode{}).Count(&stats["total"])

	// Used codes (used_count > 0)
	db.Model(&model.InviteCode{}).Where("used_count > 0").Count(&stats["used"])

	// Available codes (not expired and not fully used)
	now := time.Now()
	db.Model(&model.InviteCode{}).
		Where("(expired_at IS NULL OR expired_at > ?) AND used_count < max_uses", now).
		Count(&stats["available"])

	// Expired codes
	db.Model(&model.InviteCode{}).Where("expired_at IS NOT NULL AND expired_at <= ?", now).Count(&stats["expired"])

	return stats, nil
}
