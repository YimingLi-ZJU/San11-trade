package service

import (
	"errors"

	"san11-trade/internal/database"
	"san11-trade/internal/model"
)

var (
	ErrNotInAuctionPhase       = errors.New("not in auction phase")
	ErrAuctionGeneralNotFound  = errors.New("auction general not found")
	ErrGeneralAlreadyAuctioned = errors.New("general has already been auctioned")
	ErrAuctionRecordNotFound   = errors.New("auction record not found")
)

// GetAuctionPool returns all generals in the auction pool
func GetAuctionPool() ([]model.General, error) {
	db := database.GetDB()

	var generals []model.General
	if err := db.Where("pool_type = ?", "auction").
		Preload("Owner").
		Order("excel_id asc").
		Find(&generals).Error; err != nil {
		return nil, err
	}

	return generals, nil
}

// AuctionResult represents an auction result for display
type AuctionResult struct {
	GeneralID   uint           `json:"general_id"`
	GeneralName string         `json:"general_name"`
	Salary      int            `json:"salary"`
	UserID      *uint          `json:"user_id"`
	Nickname    string         `json:"nickname"`
	Price       int            `json:"price"`
	IsUnsold    bool           `json:"is_unsold"`
	Remark      string         `json:"remark"`
	General     *model.General `json:"general,omitempty"`
}

// GetAuctionResults returns all auction results
func GetAuctionResults() ([]AuctionResult, error) {
	db := database.GetDB()

	// Get all auction generals
	var generals []model.General
	if err := db.Where("pool_type = ?", "auction").
		Order("excel_id asc").
		Find(&generals).Error; err != nil {
		return nil, err
	}

	// Get auction records
	var records []model.AuctionRecord
	if err := db.Preload("User").Preload("General").Find(&records).Error; err != nil {
		return nil, err
	}

	// Create a map of general_id -> record
	recordMap := make(map[uint]*model.AuctionRecord)
	for i := range records {
		recordMap[records[i].GeneralID] = &records[i]
	}

	// Build results
	results := make([]AuctionResult, 0, len(generals))
	for _, general := range generals {
		result := AuctionResult{
			GeneralID:   general.ID,
			GeneralName: general.Name,
			Salary:      general.Salary,
			General:     &general,
		}

		if record, exists := recordMap[general.ID]; exists {
			result.UserID = record.UserID
			result.Price = record.Price
			result.IsUnsold = record.IsUnsold
			result.Remark = record.Remark
			if record.User != nil {
				result.Nickname = record.User.Nickname
			}
		}

		results = append(results, result)
	}

	return results, nil
}

// AssignAuctionRequest represents a request to assign auction result
type AssignAuctionRequest struct {
	GeneralID uint   `json:"general_id" binding:"required"`
	UserID    *uint  `json:"user_id"` // null means unsold (流拍)
	Price     int    `json:"price"`   // Auction price (space cost)
	Remark    string `json:"remark"`
}

// AssignAuction assigns an auction general to a user (admin only)
func AssignAuction(req *AssignAuctionRequest) (*model.AuctionRecord, error) {
	db := database.GetDB()

	// Check phase (allow admin to operate even not in auction phase for flexibility)
	// phase, err := GetGamePhase()
	// if err != nil {
	// 	return nil, err
	// }
	// if phase.CurrentPhase != "auction" {
	// 	return nil, ErrNotInAuctionPhase
	// }

	// Get the general
	var general model.General
	if err := db.First(&general, req.GeneralID).Error; err != nil {
		return nil, ErrAuctionGeneralNotFound
	}

	// Verify it's an auction general
	if general.PoolType != "auction" {
		return nil, ErrAuctionGeneralNotFound
	}

	// Check if already auctioned
	var existingRecord model.AuctionRecord
	if err := db.Where("general_id = ?", req.GeneralID).First(&existingRecord).Error; err == nil {
		return nil, ErrGeneralAlreadyAuctioned
	}

	// Determine price: use provided price, or default to salary
	price := req.Price
	if price == 0 && req.UserID != nil {
		price = general.Salary
	}

	// Begin transaction
	tx := db.Begin()

	// Create auction record
	record := &model.AuctionRecord{
		GeneralID: req.GeneralID,
		UserID:    req.UserID,
		Price:     price,
		IsUnsold:  req.UserID == nil,
		Remark:    req.Remark,
	}

	if err := tx.Create(record).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// If there's a winner, assign the general and update space
	if req.UserID != nil {
		// Update general owner
		if err := tx.Model(&general).Updates(map[string]interface{}{
			"owner_id":     *req.UserID,
			"is_available": false,
		}).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		// Update user's used space
		if err := tx.Model(&model.User{}).Where("id = ?", *req.UserID).
			UpdateColumn("used_space", database.GetDB().Raw("used_space + ?", price)).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// Reload record with associations
	db.Preload("User").Preload("General").First(record, record.ID)

	return record, nil
}

// ResetAuction resets an auction record (admin only)
func ResetAuction(recordID uint) error {
	db := database.GetDB()

	// Get the auction record
	var record model.AuctionRecord
	if err := db.First(&record, recordID).Error; err != nil {
		return ErrAuctionRecordNotFound
	}

	// Begin transaction
	tx := db.Begin()

	// If there was a winner, reverse the changes
	if record.UserID != nil {
		// Remove general owner
		if err := tx.Model(&model.General{}).Where("id = ?", record.GeneralID).Updates(map[string]interface{}{
			"owner_id":     nil,
			"is_available": true,
		}).Error; err != nil {
			tx.Rollback()
			return err
		}

		// Restore user's space
		if err := tx.Model(&model.User{}).Where("id = ?", *record.UserID).
			UpdateColumn("used_space", database.GetDB().Raw("used_space - ?", record.Price)).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Delete the auction record
	if err := tx.Delete(&record).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// ResetAuctionByGeneralID resets an auction record by general ID (admin only)
func ResetAuctionByGeneralID(generalID uint) error {
	db := database.GetDB()

	// Get the auction record
	var record model.AuctionRecord
	if err := db.Where("general_id = ?", generalID).First(&record).Error; err != nil {
		return ErrAuctionRecordNotFound
	}

	return ResetAuction(record.ID)
}

// GetAuctionStats returns auction statistics
type AuctionStats struct {
	TotalGenerals int `json:"total_generals"`
	Auctioned     int `json:"auctioned"`
	Sold          int `json:"sold"`
	Unsold        int `json:"unsold"`
	Pending       int `json:"pending"`
	TotalPrice    int `json:"total_price"`
}

func GetAuctionStats() (*AuctionStats, error) {
	db := database.GetDB()

	stats := &AuctionStats{}

	// Total auction generals
	var totalGenerals int64
	db.Model(&model.General{}).Where("pool_type = ?", "auction").Count(&totalGenerals)
	stats.TotalGenerals = int(totalGenerals)

	// Auctioned count
	var auctioned int64
	db.Model(&model.AuctionRecord{}).Count(&auctioned)
	stats.Auctioned = int(auctioned)

	// Sold count (has winner)
	var sold int64
	db.Model(&model.AuctionRecord{}).Where("user_id IS NOT NULL").Count(&sold)
	stats.Sold = int(sold)

	// Unsold count (no winner)
	var unsold int64
	db.Model(&model.AuctionRecord{}).Where("user_id IS NULL").Count(&unsold)
	stats.Unsold = int(unsold)

	// Pending (not yet auctioned)
	stats.Pending = stats.TotalGenerals - stats.Auctioned

	// Total price
	var totalPrice struct {
		Sum int64
	}
	db.Model(&model.AuctionRecord{}).Select("COALESCE(SUM(price), 0) as sum").Scan(&totalPrice)
	stats.TotalPrice = int(totalPrice.Sum)

	return stats, nil
}
