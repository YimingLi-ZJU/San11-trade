package service

import (
	"errors"

	"san11-trade/internal/database"
	"san11-trade/internal/model"
)

var (
	ErrNotInSignupPhase  = errors.New("not in signup phase")
	ErrAlreadyRegistered = errors.New("already registered for this season")
	ErrRegistrationFull  = errors.New("registration is full")
	ErrPhaseNotFound     = errors.New("game phase not found")
)

// GetGamePhase retrieves the current game phase
func GetGamePhase() (*model.GamePhase, error) {
	db := database.GetDB()

	var phase model.GamePhase
	if err := db.First(&phase).Error; err != nil {
		return nil, ErrPhaseNotFound
	}

	return &phase, nil
}

// SetGamePhase updates the current game phase (admin only)
func SetGamePhase(phaseName string, roundNumber int, draftRound int) error {
	db := database.GetDB()

	validPhases := map[string]bool{
		"signup":   true,
		"draw":     true,
		"draft":    true,
		"trading":  true,
		"auction":  true,
		"match":    true,
		"finished": true,
	}

	if !validPhases[phaseName] {
		return errors.New("invalid phase name")
	}

	return db.Model(&model.GamePhase{}).Where("id = 1").Updates(map[string]interface{}{
		"current_phase": phaseName,
		"round_number":  roundNumber,
		"draft_round":   draftRound,
	}).Error
}

// SignUp registers a user for the current season
func SignUp(userID uint) error {
	db := database.GetDB()

	// Check current phase
	phase, err := GetGamePhase()
	if err != nil {
		return err
	}

	if phase.CurrentPhase != "signup" {
		return ErrNotInSignupPhase
	}

	// Check if already registered
	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		return err
	}

	if user.IsRegistered {
		return ErrAlreadyRegistered
	}

	// Check if registration is full
	var count int64
	db.Model(&model.User{}).Where("is_registered = ?", true).Count(&count)
	if count >= 32 {
		return ErrRegistrationFull
	}

	// Register the user
	return db.Model(&user).Updates(map[string]interface{}{
		"is_registered": true,
		"space":         350,
		"used_space":    0,
	}).Error
}

// GetRegisteredPlayers returns all registered players
func GetRegisteredPlayers() ([]model.User, error) {
	db := database.GetDB()

	var users []model.User
	if err := db.Where("is_registered = ?", true).Preload("Club").Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

// GetRegisteredCount returns the number of registered players
func GetRegisteredCount() (int64, error) {
	db := database.GetDB()

	var count int64
	if err := db.Model(&model.User{}).Where("is_registered = ?", true).Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// ResetSeason resets all game data for a new season (admin only)
func ResetSeason() error {
	db := database.GetDB()

	// Begin transaction
	tx := db.Begin()

	// Reset all generals ownership
	if err := tx.Model(&model.General{}).Updates(map[string]interface{}{
		"owner_id":      nil,
		"is_available":  true,
		"injured_until": nil,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Reset all treasures ownership
	if err := tx.Model(&model.Treasure{}).Updates(map[string]interface{}{
		"owner_id":     nil,
		"is_available": true,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Reset all clubs ownership
	if err := tx.Model(&model.Club{}).Update("owner_id", nil).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Reset all users
	if err := tx.Model(&model.User{}).Where("is_admin = ?", false).Updates(map[string]interface{}{
		"is_registered": false,
		"space":         350,
		"used_space":    0,
		"club_id":       nil,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Clear player_generals and player_treasures associations
	if err := tx.Exec("DELETE FROM player_generals").Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Exec("DELETE FROM player_treasures").Error; err != nil {
		tx.Rollback()
		return err
	}

	// Clear trade records
	if err := tx.Exec("DELETE FROM trades").Error; err != nil {
		tx.Rollback()
		return err
	}

	// Clear draw records
	if err := tx.Exec("DELETE FROM draw_records").Error; err != nil {
		tx.Rollback()
		return err
	}

	// Clear draft records
	if err := tx.Exec("DELETE FROM draft_records").Error; err != nil {
		tx.Rollback()
		return err
	}

	// Clear auction records
	if err := tx.Exec("DELETE FROM auction_records").Error; err != nil {
		tx.Rollback()
		return err
	}

	// Reset game phase
	if err := tx.Model(&model.GamePhase{}).Where("id = 1").Updates(map[string]interface{}{
		"current_phase": "signup",
		"round_number":  1,
		"draft_round":   0,
		"draft_order":   "[]",
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
