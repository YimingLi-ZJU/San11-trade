package service

import (
	"san11-trade/internal/database"
	"san11-trade/internal/model"
)

// GetAllGenerals returns all generals
func GetAllGenerals() ([]model.General, error) {
	db := database.GetDB()

	var generals []model.General
	if err := db.Preload("Owner").Order("tier ASC, salary DESC").Find(&generals).Error; err != nil {
		return nil, err
	}

	return generals, nil
}

// GetGeneralByID returns a general by ID
func GetGeneralByID(id uint) (*model.General, error) {
	db := database.GetDB()

	var general model.General
	if err := db.Preload("Owner").First(&general, id).Error; err != nil {
		return nil, err
	}

	return &general, nil
}

// GetUserGenerals returns all generals owned by a user
func GetUserGenerals(userID uint) ([]model.General, error) {
	db := database.GetDB()

	var generals []model.General
	if err := db.Where("owner_id = ?", userID).Order("tier ASC, salary DESC").Find(&generals).Error; err != nil {
		return nil, err
	}

	return generals, nil
}

// GetAllTreasures returns all treasures
func GetAllTreasures() ([]model.Treasure, error) {
	db := database.GetDB()

	var treasures []model.Treasure
	if err := db.Preload("Owner").Order("type ASC, value DESC").Find(&treasures).Error; err != nil {
		return nil, err
	}

	return treasures, nil
}

// GetTreasureByID returns a treasure by ID
func GetTreasureByID(id uint) (*model.Treasure, error) {
	db := database.GetDB()

	var treasure model.Treasure
	if err := db.Preload("Owner").First(&treasure, id).Error; err != nil {
		return nil, err
	}

	return &treasure, nil
}

// GetUserTreasures returns all treasures owned by a user
func GetUserTreasures(userID uint) ([]model.Treasure, error) {
	db := database.GetDB()

	var treasures []model.Treasure
	if err := db.Where("owner_id = ?", userID).Order("type ASC, value DESC").Find(&treasures).Error; err != nil {
		return nil, err
	}

	return treasures, nil
}

// GetAllClubs returns all clubs
func GetAllClubs() ([]model.Club, error) {
	db := database.GetDB()

	var clubs []model.Club
	if err := db.Preload("Owner").Find(&clubs).Error; err != nil {
		return nil, err
	}

	return clubs, nil
}

// GetClubByID returns a club by ID
func GetClubByID(id uint) (*model.Club, error) {
	db := database.GetDB()

	var club model.Club
	if err := db.Preload("Owner").First(&club, id).Error; err != nil {
		return nil, err
	}

	return &club, nil
}

// GetUserRoster returns a user's complete roster
type Roster struct {
	User      *model.User      `json:"user"`
	Generals  []model.General  `json:"generals"`
	Treasures []model.Treasure `json:"treasures"`
	Club      *model.Club      `json:"club"`
}

func GetUserRoster(userID uint) (*Roster, error) {
	user, err := GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	generals, err := GetUserGenerals(userID)
	if err != nil {
		return nil, err
	}

	treasures, err := GetUserTreasures(userID)
	if err != nil {
		return nil, err
	}

	var club *model.Club
	if user.ClubID != nil {
		club, _ = GetClubByID(*user.ClubID)
	}

	return &Roster{
		User:      user,
		Generals:  generals,
		Treasures: treasures,
		Club:      club,
	}, nil
}

// GetStatistics returns game statistics
type Statistics struct {
	TotalPlayers      int64 `json:"total_players"`
	RegisteredPlayers int64 `json:"registered_players"`
	TotalGenerals     int64 `json:"total_generals"`
	OwnedGenerals     int64 `json:"owned_generals"`
	TotalTreasures    int64 `json:"total_treasures"`
	OwnedTreasures    int64 `json:"owned_treasures"`
	TotalTrades       int64 `json:"total_trades"`
	AcceptedTrades    int64 `json:"accepted_trades"`
}

func GetStatistics() (*Statistics, error) {
	db := database.GetDB()

	var stats Statistics

	db.Model(&model.User{}).Count(&stats.TotalPlayers)
	db.Model(&model.User{}).Where("is_registered = ?", true).Count(&stats.RegisteredPlayers)
	db.Model(&model.General{}).Count(&stats.TotalGenerals)
	db.Model(&model.General{}).Where("owner_id IS NOT NULL").Count(&stats.OwnedGenerals)
	db.Model(&model.Treasure{}).Count(&stats.TotalTreasures)
	db.Model(&model.Treasure{}).Where("owner_id IS NOT NULL").Count(&stats.OwnedTreasures)
	db.Model(&model.Trade{}).Count(&stats.TotalTrades)
	db.Model(&model.Trade{}).Where("status = ?", "accepted").Count(&stats.AcceptedTrades)

	return &stats, nil
}
