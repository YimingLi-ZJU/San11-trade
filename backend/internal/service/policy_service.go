package service

import (
	"errors"
	"sort"
	"time"

	"san11-trade/internal/database"
	"san11-trade/internal/model"

	"gorm.io/gorm"
)

// Policy auction errors
var (
	ErrNotInPolicyPhase        = errors.New("not in policy phase")
	ErrBiddingClosed           = errors.New("bidding has been closed")
	ErrBiddingNotClosed        = errors.New("bidding has not been closed yet")
	ErrSelectionNotStarted     = errors.New("selection has not started yet")
	ErrSelectionCompleted      = errors.New("selection has already completed")
	ErrPolicyNotYourTurn       = errors.New("it's not your turn to select")
	ErrClubAlreadySelected     = errors.New("this club has already been selected")
	ErrAlreadySelected         = errors.New("you have already selected a club")
	ErrPolicyInsufficientSpace = errors.New("insufficient space for bid")
	ErrInvalidTimeoutMinutes   = errors.New("timeout must be between 5 and 60 minutes")
	ErrClubNotFound            = errors.New("club not found")
)

// GetPolicyPhaseConfig retrieves or creates the policy phase config
func GetPolicyPhaseConfig() (*model.PolicyPhaseConfig, error) {
	db := database.GetDB()
	var config model.PolicyPhaseConfig
	if err := db.First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create default config
			config = model.PolicyPhaseConfig{
				Status:         "bidding",
				TimeoutMinutes: 10,
			}
			if err := db.Create(&config).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return &config, nil
}

// UpdatePolicyPhaseConfig updates the policy phase configuration
func UpdatePolicyPhaseConfig(status string, startTime *time.Time, timeoutMinutes int, currentSelector *uint, currentDeadline *time.Time) error {
	db := database.GetDB()
	config, err := GetPolicyPhaseConfig()
	if err != nil {
		return err
	}

	updates := map[string]interface{}{}
	if status != "" {
		updates["status"] = status
	}
	if startTime != nil {
		updates["start_time"] = startTime
	}
	if timeoutMinutes > 0 {
		updates["timeout_minutes"] = timeoutMinutes
	}
	updates["current_selector"] = currentSelector
	updates["current_deadline"] = currentDeadline

	return db.Model(&config).Updates(updates).Error
}

// PlacePolicyBid places or updates a bid for policy selection
func PlacePolicyBid(userID uint, bidAmount int) error {
	db := database.GetDB()

	// Check current phase
	phase, err := GetGamePhase()
	if err != nil {
		return err
	}
	if phase.CurrentPhase != "policy" {
		return ErrNotInPolicyPhase
	}

	// Check policy phase status
	config, err := GetPolicyPhaseConfig()
	if err != nil {
		return err
	}
	if config.Status != "bidding" {
		return ErrBiddingClosed
	}

	// Check user's available space
	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		return err
	}
	if !user.IsRegistered {
		return errors.New("user is not registered")
	}
	availableSpace := user.Space - user.UsedSpace
	if bidAmount > availableSpace {
		return ErrPolicyInsufficientSpace
	}

	// Create or update bid
	var bid model.PolicyBid
	if err := db.Where("user_id = ?", userID).First(&bid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			bid = model.PolicyBid{
				UserID:    userID,
				BidAmount: bidAmount,
			}
			return db.Create(&bid).Error
		}
		return err
	}

	return db.Model(&bid).Update("bid_amount", bidAmount).Error
}

// SetPolicyPreferences sets a user's preferred club order
func SetPolicyPreferences(userID uint, clubIDs []uint) error {
	db := database.GetDB()

	// Check current phase
	phase, err := GetGamePhase()
	if err != nil {
		return err
	}
	if phase.CurrentPhase != "policy" {
		return ErrNotInPolicyPhase
	}

	// Check policy phase status
	config, err := GetPolicyPhaseConfig()
	if err != nil {
		return err
	}
	if config.Status != "bidding" {
		return ErrBiddingClosed
	}

	// Validate all club IDs exist
	for _, clubID := range clubIDs {
		var club model.Club
		if err := db.First(&club, clubID).Error; err != nil {
			return ErrClubNotFound
		}
	}

	// Start transaction
	tx := db.Begin()

	// Delete existing preferences
	if err := tx.Where("user_id = ?", userID).Delete(&model.PolicyPreference{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Create new preferences
	for i, clubID := range clubIDs {
		pref := model.PolicyPreference{
			UserID:   userID,
			ClubID:   clubID,
			Priority: i + 1,
		}
		if err := tx.Create(&pref).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// GetUserPolicyBid retrieves a user's bid
func GetUserPolicyBid(userID uint) (*model.PolicyBid, error) {
	db := database.GetDB()
	var bid model.PolicyBid
	if err := db.Where("user_id = ?", userID).First(&bid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &bid, nil
}

// GetUserPolicyPreferences retrieves a user's preferences
func GetUserPolicyPreferences(userID uint) ([]model.PolicyPreference, error) {
	db := database.GetDB()
	var prefs []model.PolicyPreference
	if err := db.Where("user_id = ?", userID).Order("priority asc").Preload("Club").Find(&prefs).Error; err != nil {
		return nil, err
	}
	return prefs, nil
}

// GetAllPolicyBids retrieves all bids (admin only after bidding closed)
func GetAllPolicyBids() ([]model.PolicyBid, error) {
	db := database.GetDB()
	var bids []model.PolicyBid
	if err := db.Preload("User").Order("rank asc, bid_amount desc").Find(&bids).Error; err != nil {
		return nil, err
	}
	return bids, nil
}

// CloseBidding closes the bidding phase and calculates selection order
func CloseBidding() error {
	db := database.GetDB()

	// Check policy phase status
	config, err := GetPolicyPhaseConfig()
	if err != nil {
		return err
	}
	if config.Status != "bidding" {
		return errors.New("bidding is not open")
	}

	// Get all bids
	var bids []model.PolicyBid
	if err := db.Preload("User").Find(&bids).Error; err != nil {
		return err
	}

	// Sort by bid amount (descending), then by created_at (ascending) for ties
	sort.Slice(bids, func(i, j int) bool {
		if bids[i].BidAmount != bids[j].BidAmount {
			return bids[i].BidAmount > bids[j].BidAmount
		}
		return bids[i].CreatedAt.Before(bids[j].CreatedAt)
	})

	// Assign ranks
	tx := db.Begin()
	for i, bid := range bids {
		if err := tx.Model(&bid).Update("rank", i+1).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Update status to closed
	if err := tx.Model(&config).Update("status", "closed").Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// StartPolicySelection starts the selection phase
func StartPolicySelection(startTime time.Time, timeoutMinutes int) error {
	db := database.GetDB()

	// Validate timeout
	if timeoutMinutes < 5 || timeoutMinutes > 60 {
		return ErrInvalidTimeoutMinutes
	}

	// Check policy phase status
	config, err := GetPolicyPhaseConfig()
	if err != nil {
		return err
	}
	if config.Status != "closed" {
		return ErrBiddingNotClosed
	}

	// Get first selector (rank 1)
	var firstBid model.PolicyBid
	if err := db.Where("rank = ?", 1).First(&firstBid).Error; err != nil {
		return errors.New("no bids found")
	}

	// Calculate first deadline
	firstDeadline := startTime.Add(time.Duration(timeoutMinutes) * time.Minute)

	// Update config
	return db.Model(&config).Updates(map[string]interface{}{
		"status":           "selecting",
		"start_time":       startTime,
		"timeout_minutes":  timeoutMinutes,
		"current_selector": firstBid.UserID,
		"current_deadline": firstDeadline,
	}).Error
}

// SelectClub allows a user to select a club during their turn
func SelectClub(userID uint, clubID uint) error {
	db := database.GetDB()

	// Check policy phase status
	config, err := GetPolicyPhaseConfig()
	if err != nil {
		return err
	}
	if config.Status != "selecting" {
		return ErrSelectionNotStarted
	}

	// Check if it's this user's turn
	if config.CurrentSelector == nil || *config.CurrentSelector != userID {
		return ErrPolicyNotYourTurn
	}

	// Check if club exists
	var club model.Club
	if err := db.First(&club, clubID).Error; err != nil {
		return ErrClubNotFound
	}

	// Check if club is already selected
	var existingSelection model.PolicySelection
	if err := db.Where("club_id = ?", clubID).First(&existingSelection).Error; err == nil {
		return ErrClubAlreadySelected
	}

	// Check if user already has a selection
	if err := db.Where("user_id = ?", userID).First(&existingSelection).Error; err == nil {
		return ErrAlreadySelected
	}

	// Get user's bid to determine cost
	var bid model.PolicyBid
	var bidCost int
	if err := db.Where("user_id = ?", userID).First(&bid).Error; err == nil {
		bidCost = bid.BidAmount
	}

	// Get current selection count for order
	var selectionCount int64
	db.Model(&model.PolicySelection{}).Count(&selectionCount)

	// Start transaction
	tx := db.Begin()

	// Create selection
	selection := model.PolicySelection{
		UserID:       userID,
		ClubID:       clubID,
		BidCost:      bidCost,
		AutoAssigned: false,
		SelectOrder:  int(selectionCount) + 1,
	}
	if err := tx.Create(&selection).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Update user's club and used space
	if err := tx.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"club_id":    clubID,
		"used_space": gorm.Expr("used_space + ?", bidCost),
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Update club owner
	if err := tx.Model(&club).Update("owner_id", userID).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Move to next selector
	if err := moveToNextSelector(tx, config); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// moveToNextSelector moves to the next person in the selection queue
func moveToNextSelector(tx *gorm.DB, config *model.PolicyPhaseConfig) error {
	// Get all bids ordered by rank
	var bids []model.PolicyBid
	if err := tx.Order("rank asc").Find(&bids).Error; err != nil {
		return err
	}

	// Find the next user who hasn't selected yet
	for _, bid := range bids {
		var selection model.PolicySelection
		if err := tx.Where("user_id = ?", bid.UserID).First(&selection).Error; err == gorm.ErrRecordNotFound {
			// This user hasn't selected yet, they're next
			deadline := time.Now().Add(time.Duration(config.TimeoutMinutes) * time.Minute)
			return tx.Model(&config).Updates(map[string]interface{}{
				"current_selector": bid.UserID,
				"current_deadline": deadline,
			}).Error
		}
	}

	// Everyone has selected, complete the phase
	return tx.Model(&config).Updates(map[string]interface{}{
		"status":           "completed",
		"current_selector": nil,
		"current_deadline": nil,
	}).Error
}

// AutoAssignClub auto-assigns a club to a user who timed out
func AutoAssignClub(userID uint) error {
	db := database.GetDB()

	// Get user's preferences
	var prefs []model.PolicyPreference
	db.Where("user_id = ?", userID).Order("priority asc").Find(&prefs)

	// Try to find an available club based on preferences
	var selectedClubID uint
	for _, pref := range prefs {
		var existingSelection model.PolicySelection
		if err := db.Where("club_id = ?", pref.ClubID).First(&existingSelection).Error; err == gorm.ErrRecordNotFound {
			// This club is available
			selectedClubID = pref.ClubID
			break
		}
	}

	// If no preferred club is available, find any available club
	if selectedClubID == 0 {
		var clubs []model.Club
		db.Find(&clubs)
		for _, club := range clubs {
			var existingSelection model.PolicySelection
			if err := db.Where("club_id = ?", club.ID).First(&existingSelection).Error; err == gorm.ErrRecordNotFound {
				selectedClubID = club.ID
				break
			}
		}
	}

	if selectedClubID == 0 {
		return errors.New("no available clubs")
	}

	// Get user's bid to determine cost
	var bid model.PolicyBid
	var bidCost int
	if err := db.Where("user_id = ?", userID).First(&bid).Error; err == nil {
		bidCost = bid.BidAmount
	}

	// Get current selection count for order
	var selectionCount int64
	db.Model(&model.PolicySelection{}).Count(&selectionCount)

	// Get config
	config, err := GetPolicyPhaseConfig()
	if err != nil {
		return err
	}

	// Start transaction
	tx := db.Begin()

	// Create auto-assigned selection
	selection := model.PolicySelection{
		UserID:       userID,
		ClubID:       selectedClubID,
		BidCost:      bidCost,
		AutoAssigned: true,
		SelectOrder:  int(selectionCount) + 1,
	}
	if err := tx.Create(&selection).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Update user's club and used space
	if err := tx.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"club_id":    selectedClubID,
		"used_space": gorm.Expr("used_space + ?", bidCost),
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Update club owner
	if err := tx.Model(&model.Club{}).Where("id = ?", selectedClubID).Update("owner_id", userID).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Move to next selector
	if err := moveToNextSelector(tx, config); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// CheckAndHandleTimeout checks if current selector has timed out and handles it
func CheckAndHandleTimeout() (bool, error) {
	config, err := GetPolicyPhaseConfig()
	if err != nil {
		return false, err
	}

	if config.Status != "selecting" || config.CurrentSelector == nil || config.CurrentDeadline == nil {
		return false, nil
	}

	// Check if the start time has passed (official selection period)
	if config.StartTime != nil && time.Now().Before(*config.StartTime) {
		// Selection hasn't officially started yet, no timeout enforcement
		return false, nil
	}

	// Check if deadline has passed
	if time.Now().After(*config.CurrentDeadline) {
		// Timeout! Auto-assign
		if err := AutoAssignClub(*config.CurrentSelector); err != nil {
			return false, err
		}
		return true, nil
	}

	return false, nil
}

// GetPolicySelectionStatus returns the current selection status
func GetPolicySelectionStatus() (map[string]interface{}, error) {
	db := database.GetDB()

	config, err := GetPolicyPhaseConfig()
	if err != nil {
		return nil, err
	}

	// Get all bids with rank
	var bids []model.PolicyBid
	db.Preload("User").Order("rank asc").Find(&bids)

	// Get all selections
	var selections []model.PolicySelection
	db.Preload("User").Preload("Club").Order("select_order asc").Find(&selections)

	// Get available clubs
	var selectedClubIDs []uint
	for _, sel := range selections {
		selectedClubIDs = append(selectedClubIDs, sel.ClubID)
	}

	var availableClubs []model.Club
	if len(selectedClubIDs) > 0 {
		db.Where("id NOT IN ?", selectedClubIDs).Preload("Tags").Preload("Policies").Find(&availableClubs)
	} else {
		db.Preload("Tags").Preload("Policies").Find(&availableClubs)
	}

	// Get current selector info
	var currentUser *model.User
	if config.CurrentSelector != nil {
		currentUser = &model.User{}
		db.First(currentUser, *config.CurrentSelector)
	}

	return map[string]interface{}{
		"config":          config,
		"bids":            bids,
		"selections":      selections,
		"available_clubs": availableClubs,
		"current_user":    currentUser,
	}, nil
}

// GetAllPolicySelections retrieves all selections
func GetAllPolicySelections() ([]model.PolicySelection, error) {
	db := database.GetDB()
	var selections []model.PolicySelection
	if err := db.Preload("User").Preload("Club").Order("select_order asc").Find(&selections).Error; err != nil {
		return nil, err
	}
	return selections, nil
}

// ResetPolicyPhase resets the entire policy phase (admin only)
func ResetPolicyPhase() error {
	db := database.GetDB()
	tx := db.Begin()

	// Delete all selections
	if err := tx.Exec("DELETE FROM policy_selections").Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete all preferences
	if err := tx.Exec("DELETE FROM policy_preferences").Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete all bids
	if err := tx.Exec("DELETE FROM policy_bids").Error; err != nil {
		tx.Rollback()
		return err
	}

	// Reset all users' club assignments
	if err := tx.Model(&model.User{}).Where("club_id IS NOT NULL").Updates(map[string]interface{}{
		"club_id": nil,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Reset all clubs' owner assignments
	if err := tx.Model(&model.Club{}).Where("owner_id IS NOT NULL").Update("owner_id", nil).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Reset policy phase config
	var config model.PolicyPhaseConfig
	if err := tx.First(&config).Error; err == nil {
		if err := tx.Model(&config).Updates(map[string]interface{}{
			"status":           "bidding",
			"start_time":       nil,
			"current_selector": nil,
			"current_deadline": nil,
		}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// ResetUserPolicySelection resets a specific user's selection (admin only)
func ResetUserPolicySelection(userID uint) error {
	db := database.GetDB()

	// Get the selection
	var selection model.PolicySelection
	if err := db.Where("user_id = ?", userID).First(&selection).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("user has not selected a club")
		}
		return err
	}

	tx := db.Begin()

	// Reset user's club and refund space
	if err := tx.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"club_id":    nil,
		"used_space": gorm.Expr("used_space - ?", selection.BidCost),
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Reset club owner
	if err := tx.Model(&model.Club{}).Where("id = ?", selection.ClubID).Update("owner_id", nil).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete selection
	if err := tx.Delete(&selection).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// GetClubsWithTags retrieves all clubs with their tags
func GetClubsWithTags() ([]model.Club, error) {
	db := database.GetDB()
	var clubs []model.Club
	if err := db.Preload("Tags").Preload("Policies").Preload("Owner").Order("excel_id asc").Find(&clubs).Error; err != nil {
		return nil, err
	}
	return clubs, nil
}

// GetClubsByTag retrieves clubs with a specific tag
func GetClubsByTag(tag string) ([]model.Club, error) {
	db := database.GetDB()
	var clubTags []model.ClubTag
	if err := db.Where("tag = ?", tag).Find(&clubTags).Error; err != nil {
		return nil, err
	}

	var clubIDs []uint
	for _, ct := range clubTags {
		clubIDs = append(clubIDs, ct.ClubID)
	}

	if len(clubIDs) == 0 {
		return []model.Club{}, nil
	}

	var clubs []model.Club
	if err := db.Where("id IN ?", clubIDs).Preload("Tags").Preload("Policies").Preload("Owner").Find(&clubs).Error; err != nil {
		return nil, err
	}
	return clubs, nil
}

// GetClubsByLeague retrieves clubs in a specific league
func GetClubsByLeague(league string) ([]model.Club, error) {
	db := database.GetDB()
	var clubs []model.Club
	if err := db.Where("league = ?", league).Preload("Tags").Preload("Policies").Preload("Owner").Find(&clubs).Error; err != nil {
		return nil, err
	}
	return clubs, nil
}

// GetAllLeagues retrieves all unique leagues
func GetAllLeagues() ([]string, error) {
	db := database.GetDB()
	var leagues []string
	if err := db.Model(&model.Club{}).Distinct("league").Where("league != ''").Pluck("league", &leagues).Error; err != nil {
		return nil, err
	}
	return leagues, nil
}

// GetAllTags retrieves all unique tags
func GetAllTags() ([]string, error) {
	db := database.GetDB()
	var tags []string
	if err := db.Model(&model.ClubTag{}).Distinct("tag").Pluck("tag", &tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}
