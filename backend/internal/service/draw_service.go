package service

import (
	"errors"
	"math/rand"
	"time"

	"san11-trade/internal/config"
	"san11-trade/internal/database"
	"san11-trade/internal/model"
)

var (
	ErrNotInDrawPhase      = errors.New("not in draw phase")
	ErrDrawLimitReached    = errors.New("draw limit reached")
	ErrNoAvailableGenerals = errors.New("no available generals in pool")
	ErrNotInDraftPhase     = errors.New("not in draft phase")
	ErrNotYourTurn         = errors.New("not your turn to draft")
	ErrGeneralNotAvailable = errors.New("general not available for draft")
	ErrInsufficientSpace   = errors.New("insufficient space for this general")
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// GetDrawCount returns the number of draws a user has made
func GetDrawCount(userID uint, drawType string) (int, error) {
	db := database.GetDB()

	var count int64
	if err := db.Model(&model.DrawRecord{}).
		Where("user_id = ? AND draw_type = ?", userID, drawType).
		Count(&count).Error; err != nil {
		return 0, err
	}

	return int(count), nil
}

// GuaranteeDraw performs a guarantee draw for a user
func GuaranteeDraw(userID uint) (*model.General, error) {
	db := database.GetDB()

	// Check phase
	phase, err := GetGamePhase()
	if err != nil {
		return nil, err
	}
	if phase.CurrentPhase != "guarantee_draw" {
		return nil, ErrNotInDrawPhase
	}

	// Check draw count
	count, err := GetDrawCount(userID, "guarantee")
	if err != nil {
		return nil, err
	}
	if count >= config.AppConfig.Game.GuaranteeDraws {
		return nil, ErrDrawLimitReached
	}

	// Get user for space check
	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		return nil, err
	}

	// Get available generals from guarantee pool
	var generals []model.General
	if err := db.Where("pool_type = ? AND is_available = ? AND owner_id IS NULL", "guarantee", true).
		Find(&generals).Error; err != nil {
		return nil, err
	}

	if len(generals) == 0 {
		return nil, ErrNoAvailableGenerals
	}

	// Random select one
	selected := generals[rand.Intn(len(generals))]

	// Check space
	remainingSpace := user.Space - user.UsedSpace
	if selected.Salary > remainingSpace {
		return nil, ErrInsufficientSpace
	}

	// Begin transaction
	tx := db.Begin()

	// Assign general to user
	if err := tx.Model(&selected).Updates(map[string]interface{}{
		"owner_id":     userID,
		"is_available": false,
	}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Update user's used space
	if err := tx.Model(&user).Update("used_space", user.UsedSpace+selected.Salary).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Record the draw
	record := model.DrawRecord{
		UserID:    userID,
		GeneralID: selected.ID,
		DrawType:  "guarantee",
	}
	if err := tx.Create(&record).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &selected, nil
}

// NormalDraw performs a normal draw for a user
func NormalDraw(userID uint) (*model.General, error) {
	db := database.GetDB()

	// Check phase
	phase, err := GetGamePhase()
	if err != nil {
		return nil, err
	}
	if phase.CurrentPhase != "normal_draw" {
		return nil, ErrNotInDrawPhase
	}

	// Check draw count
	count, err := GetDrawCount(userID, "normal")
	if err != nil {
		return nil, err
	}
	if count >= config.AppConfig.Game.NormalDraws {
		return nil, ErrDrawLimitReached
	}

	// Get user for space check
	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		return nil, err
	}

	// Get available generals from normal pool
	var generals []model.General
	if err := db.Where("pool_type = ? AND is_available = ? AND owner_id IS NULL", "normal", true).
		Find(&generals).Error; err != nil {
		return nil, err
	}

	if len(generals) == 0 {
		return nil, ErrNoAvailableGenerals
	}

	// Random select one
	selected := generals[rand.Intn(len(generals))]

	// Check space
	remainingSpace := user.Space - user.UsedSpace
	if selected.Salary > remainingSpace {
		return nil, ErrInsufficientSpace
	}

	// Begin transaction
	tx := db.Begin()

	// Assign general to user
	if err := tx.Model(&selected).Updates(map[string]interface{}{
		"owner_id":     userID,
		"is_available": false,
	}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Update user's used space
	if err := tx.Model(&user).Update("used_space", user.UsedSpace+selected.Salary).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Record the draw
	record := model.DrawRecord{
		UserID:    userID,
		GeneralID: selected.ID,
		DrawType:  "normal",
	}
	if err := tx.Create(&record).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &selected, nil
}

// GetAvailablePoolGenerals returns available generals in a pool
func GetAvailablePoolGenerals(poolType string) ([]model.General, error) {
	db := database.GetDB()

	var generals []model.General
	if err := db.Where("pool_type = ? AND is_available = ? AND owner_id IS NULL", poolType, true).
		Find(&generals).Error; err != nil {
		return nil, err
	}

	return generals, nil
}

// GetDraftPool returns available generals for draft
func GetDraftPool() ([]model.General, error) {
	db := database.GetDB()

	var generals []model.General
	if err := db.Where("pool_type = ? AND is_available = ? AND owner_id IS NULL", "draft", true).
		Find(&generals).Error; err != nil {
		return nil, err
	}

	return generals, nil
}

// DraftPick performs a draft pick for a user
func DraftPick(userID uint, generalID uint) (*model.General, error) {
	db := database.GetDB()

	// Check phase
	phase, err := GetGamePhase()
	if err != nil {
		return nil, err
	}
	if phase.CurrentPhase != "draft" {
		return nil, ErrNotInDraftPhase
	}

	// Get user
	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		return nil, err
	}

	// Get general
	var general model.General
	if err := db.First(&general, generalID).Error; err != nil {
		return nil, err
	}

	// Check if general is available
	if general.OwnerID != nil || !general.IsAvailable || general.PoolType != "draft" {
		return nil, ErrGeneralNotAvailable
	}

	// Check space
	remainingSpace := user.Space - user.UsedSpace
	if general.Salary > remainingSpace {
		return nil, ErrInsufficientSpace
	}

	// Begin transaction
	tx := db.Begin()

	// Assign general to user
	if err := tx.Model(&general).Updates(map[string]interface{}{
		"owner_id":     userID,
		"is_available": false,
	}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Update user's used space
	if err := tx.Model(&user).Update("used_space", user.UsedSpace+general.Salary).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Record the draft
	record := model.DraftRecord{
		UserID:    userID,
		GeneralID: general.ID,
		Round:     phase.DraftRound,
	}
	if err := tx.Create(&record).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &general, nil
}

// GetUserDrawRecords returns draw records for a user
func GetUserDrawRecords(userID uint) ([]model.DrawRecord, error) {
	db := database.GetDB()

	var records []model.DrawRecord
	if err := db.Where("user_id = ?", userID).Preload("General").Find(&records).Error; err != nil {
		return nil, err
	}

	return records, nil
}

// GetUserDraftRecords returns draft records for a user
func GetUserDraftRecords(userID uint) ([]model.DraftRecord, error) {
	db := database.GetDB()

	var records []model.DraftRecord
	if err := db.Where("user_id = ?", userID).Preload("General").Find(&records).Error; err != nil {
		return nil, err
	}

	return records, nil
}
