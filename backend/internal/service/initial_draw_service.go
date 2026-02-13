package service

import (
	"errors"
	"math/rand"

	"san11-trade/internal/database"
	"san11-trade/internal/model"
)

const (
	// GuaranteeDraws is the number of guarantee draws
	GuaranteeDraws = 3
	// NormalDraws is the number of normal draws
	NormalDraws = 7
)

var (
	ErrNotInDrawPhase         = errors.New("not in draw phase")
	ErrDrawLimitReached       = errors.New("draw limit reached")
	ErrNoAvailableDrawGeneral = errors.New("no available generals in draw pool")
	ErrUserNotRegistered      = errors.New("user not registered")
)

// GetDrawCount returns the count of draws for a user by draw type
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

// DrawStatus represents the status of draws for a user
type DrawStatus struct {
	GuaranteeRemaining int `json:"guarantee_remaining"`
	NormalRemaining    int `json:"normal_remaining"`
	TotalRemaining     int `json:"total_remaining"`
	GuaranteeDone      int `json:"guarantee_done"`
	NormalDone         int `json:"normal_done"`
	TotalDone          int `json:"total_done"`
}

// GetDrawStatus returns the draw status for a user
func GetDrawStatus(userID uint) (*DrawStatus, error) {
	guaranteeDone, err := GetDrawCount(userID, "initial_guarantee")
	if err != nil {
		return nil, err
	}
	normalDone, err := GetDrawCount(userID, "initial_normal")
	if err != nil {
		return nil, err
	}

	return &DrawStatus{
		GuaranteeRemaining: GuaranteeDraws - guaranteeDone,
		NormalRemaining:    NormalDraws - normalDone,
		TotalRemaining:     (GuaranteeDraws + NormalDraws) - (guaranteeDone + normalDone),
		GuaranteeDone:      guaranteeDone,
		NormalDone:         normalDone,
		TotalDone:          guaranteeDone + normalDone,
	}, nil
}

// Draw performs a draw for a user
// It automatically picks from guarantee pool first, then normal pool
func Draw(userID uint) (*model.General, string, error) {
	// Check phase
	phase, err := GetGamePhase()
	if err != nil {
		return nil, "", err
	}
	if phase.CurrentPhase != "draw" {
		return nil, "", ErrNotInDrawPhase
	}

	return performDraw(userID)
}

// AdminDraw performs a draw for a user (admin only, no phase check)
func AdminDraw(userID uint) (*model.General, string, error) {
	return performDraw(userID)
}

// performDraw is the core draw logic
func performDraw(userID uint) (*model.General, string, error) {
	db := database.GetDB()
	// Get current draw status
	status, err := GetDrawStatus(userID)
	if err != nil {
		return nil, "", err
	}

	// Check if all draws are done
	if status.TotalRemaining <= 0 {
		return nil, "", ErrDrawLimitReached
	}

	// Get user for space check
	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		return nil, "", ErrUserNotFound
	}

	if !user.IsRegistered {
		return nil, "", ErrUserNotRegistered
	}

	// Determine which pool to draw from
	var poolType string
	var drawType string
	if status.GuaranteeRemaining > 0 {
		poolType = "initial_guarantee"
		drawType = "initial_guarantee"
	} else {
		poolType = "initial_normal"
		drawType = "initial_normal"
	}

	// Get available generals from the pool
	var generals []model.General
	if err := db.Where("pool_type = ? AND is_available = ? AND owner_id IS NULL", poolType, true).
		Find(&generals).Error; err != nil {
		return nil, "", err
	}

	if len(generals) == 0 {
		return nil, "", ErrNoAvailableDrawGeneral
	}

	// Random select one
	selected := generals[rand.Intn(len(generals))]

	// Check space
	remainingSpace := user.Space - user.UsedSpace
	if selected.Salary > remainingSpace {
		return nil, "", ErrInsufficientSpace
	}

	// Begin transaction
	tx := db.Begin()

	// Assign general to user
	if err := tx.Model(&selected).Updates(map[string]interface{}{
		"owner_id":     userID,
		"is_available": false,
	}).Error; err != nil {
		tx.Rollback()
		return nil, "", err
	}

	// Update user's used space
	if err := tx.Model(&user).Update("used_space", user.UsedSpace+selected.Salary).Error; err != nil {
		tx.Rollback()
		return nil, "", err
	}

	// Record the draw
	record := model.DrawRecord{
		UserID:    userID,
		GeneralID: selected.ID,
		DrawType:  drawType,
	}
	if err := tx.Create(&record).Error; err != nil {
		tx.Rollback()
		return nil, "", err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, "", err
	}

	return &selected, drawType, nil
}

// DrawResult represents a player's draw result
type DrawResult struct {
	UserID       uint            `json:"user_id"`
	Nickname     string          `json:"nickname"`
	Generals     []model.General `json:"generals"`
	TotalSalary  int             `json:"total_salary"`
	DrawComplete bool            `json:"draw_complete"`
}

// GetAllDrawResults returns all players' draw results
func GetAllDrawResults() ([]DrawResult, error) {
	db := database.GetDB()

	// Get all registered users
	var users []model.User
	if err := db.Where("is_registered = ?", true).Find(&users).Error; err != nil {
		return nil, err
	}

	results := make([]DrawResult, 0, len(users))
	for _, user := range users {
		// Get draw records for this user
		var records []model.DrawRecord
		if err := db.Where("user_id = ? AND (draw_type = ? OR draw_type = ?)",
			user.ID, "initial_guarantee", "initial_normal").
			Preload("General").
			Find(&records).Error; err != nil {
			return nil, err
		}

		generals := make([]model.General, 0, len(records))
		totalSalary := 0
		for _, record := range records {
			generals = append(generals, record.General)
			totalSalary += record.General.Salary
		}

		results = append(results, DrawResult{
			UserID:       user.ID,
			Nickname:     user.Nickname,
			Generals:     generals,
			TotalSalary:  totalSalary,
			DrawComplete: len(records) >= GuaranteeDraws+NormalDraws,
		})
	}

	return results, nil
}

// GetDrawPool returns available generals in draw pools
func GetDrawPool(poolType string) ([]model.General, error) {
	db := database.GetDB()

	var generals []model.General
	if err := db.Where("pool_type = ? AND is_available = ? AND owner_id IS NULL", poolType, true).
		Find(&generals).Error; err != nil {
		return nil, err
	}

	return generals, nil
}

// ResetUserDraw resets a user's draw results (admin only)
func ResetUserDraw(userID uint) error {
	db := database.GetDB()

	// Get user
	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		return ErrUserNotFound
	}

	// Get all draw records for this user
	var records []model.DrawRecord
	if err := db.Where("user_id = ? AND (draw_type = ? OR draw_type = ?)",
		userID, "initial_guarantee", "initial_normal").
		Preload("General").
		Find(&records).Error; err != nil {
		return err
	}

	if len(records) == 0 {
		return nil // Nothing to reset
	}

	// Begin transaction
	tx := db.Begin()

	// Calculate total salary to return
	totalSalary := 0
	for _, record := range records {
		totalSalary += record.General.Salary

		// Return general to pool
		if err := tx.Model(&model.General{}).Where("id = ?", record.GeneralID).Updates(map[string]interface{}{
			"owner_id":     nil,
			"is_available": true,
		}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Delete draw records
	if err := tx.Where("user_id = ? AND (draw_type = ? OR draw_type = ?)",
		userID, "initial_guarantee", "initial_normal").
		Delete(&model.DrawRecord{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Update user's used space
	newUsedSpace := user.UsedSpace - totalSalary
	if newUsedSpace < 0 {
		newUsedSpace = 0
	}
	if err := tx.Model(&user).Update("used_space", newUsedSpace).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// ResetAllUsersDraw resets all users' draw results (admin only)
func ResetAllUsersDraw() (int, error) {
	db := database.GetDB()

	// Get all registered users
	var users []model.User
	if err := db.Where("is_registered = ?", true).Find(&users).Error; err != nil {
		return 0, err
	}

	resetCount := 0
	for _, user := range users {
		if err := ResetUserDraw(user.ID); err != nil {
			return resetCount, err
		}
		resetCount++
	}

	return resetCount, nil
}

// DrawForUser performs all draws for a user (admin only)
func DrawForUser(userID uint) ([]model.General, error) {
	generals := make([]model.General, 0)

	for {
		general, _, err := AdminDraw(userID)
		if err == ErrDrawLimitReached {
			break // Done
		}
		if err != nil {
			return generals, err
		}
		generals = append(generals, *general)
	}

	return generals, nil
}

// DrawForAllUsers performs all draws for all registered users (admin only)
func DrawForAllUsers() (map[uint][]model.General, error) {
	db := database.GetDB()

	// Get all registered users
	var users []model.User
	if err := db.Where("is_registered = ?", true).Find(&users).Error; err != nil {
		return nil, err
	}

	results := make(map[uint][]model.General)
	for _, user := range users {
		generals, err := DrawForUser(user.ID)
		if err != nil && err != ErrDrawLimitReached {
			return results, err
		}
		results[user.ID] = generals
	}

	return results, nil
}

// GetInitialDrawPool returns available generals in initial draw pools
func GetInitialDrawPool(poolType string) ([]model.General, error) {
	db := database.GetDB()

	var generals []model.General
	if err := db.Where("pool_type = ? AND is_available = ? AND owner_id IS NULL", poolType, true).
		Find(&generals).Error; err != nil {
		return nil, err
	}

	return generals, nil
}
