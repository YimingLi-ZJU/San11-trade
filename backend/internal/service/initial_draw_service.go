package service

import (
	"errors"
	"math/rand"

	"san11-trade/internal/database"
	"san11-trade/internal/model"
)

const (
	// InitialGuaranteeDraws is the number of guarantee draws in initial draw phase
	InitialGuaranteeDraws = 3
	// InitialNormalDraws is the number of normal draws in initial draw phase
	InitialNormalDraws = 7
)

var (
	ErrNotInInitialDrawPhase     = errors.New("not in initial draw phase")
	ErrInitialDrawLimitReached   = errors.New("initial draw limit reached")
	ErrNoAvailableInitialGeneral = errors.New("no available generals in initial pool")
)

// GetInitialDrawCount returns the count of initial draws for a user
func GetInitialDrawCount(userID uint, drawType string) (int, error) {
	db := database.GetDB()
	var count int64
	if err := db.Model(&model.DrawRecord{}).
		Where("user_id = ? AND draw_type = ?", userID, drawType).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

// InitialDrawStatus represents the status of initial draws for a user
type InitialDrawStatus struct {
	GuaranteeRemaining int `json:"guarantee_remaining"`
	NormalRemaining    int `json:"normal_remaining"`
	TotalRemaining     int `json:"total_remaining"`
	GuaranteeDone      int `json:"guarantee_done"`
	NormalDone         int `json:"normal_done"`
	TotalDone          int `json:"total_done"`
}

// GetInitialDrawStatus returns the initial draw status for a user
func GetInitialDrawStatus(userID uint) (*InitialDrawStatus, error) {
	guaranteeDone, err := GetInitialDrawCount(userID, "initial_guarantee")
	if err != nil {
		return nil, err
	}
	normalDone, err := GetInitialDrawCount(userID, "initial_normal")
	if err != nil {
		return nil, err
	}

	return &InitialDrawStatus{
		GuaranteeRemaining: InitialGuaranteeDraws - guaranteeDone,
		NormalRemaining:    InitialNormalDraws - normalDone,
		TotalRemaining:     (InitialGuaranteeDraws + InitialNormalDraws) - (guaranteeDone + normalDone),
		GuaranteeDone:      guaranteeDone,
		NormalDone:         normalDone,
		TotalDone:          guaranteeDone + normalDone,
	}, nil
}

// InitialDraw performs an initial draw for a user
// It automatically picks from guarantee pool first, then normal pool
func InitialDraw(userID uint) (*model.General, string, error) {
	db := database.GetDB()

	// Check phase
	phase, err := GetGamePhase()
	if err != nil {
		return nil, "", err
	}
	if phase.CurrentPhase != "initial_draw" {
		return nil, "", ErrNotInInitialDrawPhase
	}

	// Get current draw status
	status, err := GetInitialDrawStatus(userID)
	if err != nil {
		return nil, "", err
	}

	// Check if all draws are done
	if status.TotalRemaining <= 0 {
		return nil, "", ErrInitialDrawLimitReached
	}

	// Get user for space check
	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		return nil, "", err
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
		return nil, "", ErrNoAvailableInitialGeneral
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

// InitialDrawResult represents a player's initial draw result
type InitialDrawResult struct {
	UserID       uint            `json:"user_id"`
	Nickname     string          `json:"nickname"`
	Generals     []model.General `json:"generals"`
	TotalSalary  int             `json:"total_salary"`
	DrawComplete bool            `json:"draw_complete"`
}

// GetAllInitialDrawResults returns all players' initial draw results
func GetAllInitialDrawResults() ([]InitialDrawResult, error) {
	db := database.GetDB()

	// Get all registered users
	var users []model.User
	if err := db.Where("is_registered = ?", true).Find(&users).Error; err != nil {
		return nil, err
	}

	results := make([]InitialDrawResult, 0, len(users))
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

		results = append(results, InitialDrawResult{
			UserID:       user.ID,
			Nickname:     user.Nickname,
			Generals:     generals,
			TotalSalary:  totalSalary,
			DrawComplete: len(records) >= InitialGuaranteeDraws+InitialNormalDraws,
		})
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
