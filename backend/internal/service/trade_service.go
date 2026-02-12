package service

import (
	"encoding/json"
	"errors"

	"san11-trade/internal/database"
	"san11-trade/internal/model"
)

var (
	ErrNotInTradingPhase     = errors.New("not in trading phase")
	ErrCannotTradeWithSelf   = errors.New("cannot trade with yourself")
	ErrTradeNotFound         = errors.New("trade not found")
	ErrNotTradeParticipant   = errors.New("you are not a participant of this trade")
	ErrTradeAlreadyProcessed = errors.New("trade has already been processed")
	ErrInvalidTradeItems     = errors.New("invalid trade items")
	ErrItemNotOwned          = errors.New("you don't own the item you're offering")
)

// TradeItem represents an item in trade
type TradeItem struct {
	Type string `json:"type"` // general/treasure/space
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// TradeRequest represents a trade proposal request
type TradeRequest struct {
	ReceiverID       uint   `json:"receiver_id"`
	OfferGenerals    []uint `json:"offer_generals"`
	OfferTreasures   []uint `json:"offer_treasures"`
	OfferSpace       int    `json:"offer_space"`
	RequestGenerals  []uint `json:"request_generals"`
	RequestTreasures []uint `json:"request_treasures"`
	RequestSpace     int    `json:"request_space"`
	Message          string `json:"message"`
}

// CreateTrade creates a new trade proposal
func CreateTrade(proposerID uint, req *TradeRequest) (*model.Trade, error) {
	db := database.GetDB()

	// Check phase
	phase, err := GetGamePhase()
	if err != nil {
		return nil, err
	}
	if phase.CurrentPhase != "trading" && phase.CurrentPhase != "draft" {
		return nil, ErrNotInTradingPhase
	}

	// Cannot trade with self
	if proposerID == req.ReceiverID {
		return nil, ErrCannotTradeWithSelf
	}

	// Validate that proposer owns the offered items
	if err := validateOwnership(proposerID, req.OfferGenerals, req.OfferTreasures); err != nil {
		return nil, err
	}

	// Validate that receiver owns the requested items
	if err := validateOwnership(req.ReceiverID, req.RequestGenerals, req.RequestTreasures); err != nil {
		return nil, ErrInvalidTradeItems
	}

	// Serialize arrays to JSON
	offerGeneralsJSON, _ := json.Marshal(req.OfferGenerals)
	offerTreasuresJSON, _ := json.Marshal(req.OfferTreasures)
	requestGeneralsJSON, _ := json.Marshal(req.RequestGenerals)
	requestTreasuresJSON, _ := json.Marshal(req.RequestTreasures)

	trade := &model.Trade{
		ProposerID:       proposerID,
		ReceiverID:       req.ReceiverID,
		OfferGenerals:    string(offerGeneralsJSON),
		OfferTreasures:   string(offerTreasuresJSON),
		OfferSpace:       req.OfferSpace,
		RequestGenerals:  string(requestGeneralsJSON),
		RequestTreasures: string(requestTreasuresJSON),
		RequestSpace:     req.RequestSpace,
		Status:           "pending",
		Message:          req.Message,
	}

	if err := db.Create(trade).Error; err != nil {
		return nil, err
	}

	// Log the trade creation
	logTrade(trade.ID, "created", proposerID, "Trade created")

	return trade, nil
}

// validateOwnership checks if a user owns the specified generals and treasures
func validateOwnership(userID uint, generalIDs []uint, treasureIDs []uint) error {
	db := database.GetDB()

	// Check generals
	for _, gid := range generalIDs {
		var general model.General
		if err := db.First(&general, gid).Error; err != nil {
			return ErrInvalidTradeItems
		}
		if general.OwnerID == nil || *general.OwnerID != userID {
			return ErrItemNotOwned
		}
	}

	// Check treasures
	for _, tid := range treasureIDs {
		var treasure model.Treasure
		if err := db.First(&treasure, tid).Error; err != nil {
			return ErrInvalidTradeItems
		}
		if treasure.OwnerID == nil || *treasure.OwnerID != userID {
			return ErrItemNotOwned
		}
	}

	return nil
}

// AcceptTrade accepts a trade proposal
func AcceptTrade(tradeID uint, userID uint) error {
	db := database.GetDB()

	var trade model.Trade
	if err := db.First(&trade, tradeID).Error; err != nil {
		return ErrTradeNotFound
	}

	// Check if user is the receiver
	if trade.ReceiverID != userID {
		return ErrNotTradeParticipant
	}

	// Check if trade is still pending
	if trade.Status != "pending" {
		return ErrTradeAlreadyProcessed
	}

	// Parse items
	var offerGenerals, requestGenerals []uint
	var offerTreasures, requestTreasures []uint
	json.Unmarshal([]byte(trade.OfferGenerals), &offerGenerals)
	json.Unmarshal([]byte(trade.OfferTreasures), &offerTreasures)
	json.Unmarshal([]byte(trade.RequestGenerals), &requestGenerals)
	json.Unmarshal([]byte(trade.RequestTreasures), &requestTreasures)

	// Re-validate ownership before executing
	if err := validateOwnership(trade.ProposerID, offerGenerals, offerTreasures); err != nil {
		// Items no longer valid, cancel trade
		db.Model(&trade).Update("status", "cancelled")
		return errors.New("proposer no longer owns the offered items")
	}
	if err := validateOwnership(trade.ReceiverID, requestGenerals, requestTreasures); err != nil {
		db.Model(&trade).Update("status", "cancelled")
		return errors.New("receiver no longer owns the requested items")
	}

	// Begin transaction
	tx := db.Begin()

	// Get users for space calculations
	var proposer, receiver model.User
	tx.First(&proposer, trade.ProposerID)
	tx.First(&receiver, trade.ReceiverID)

	// Calculate space changes
	var proposerSpaceChange, receiverSpaceChange int

	// Transfer generals from proposer to receiver
	for _, gid := range offerGenerals {
		var g model.General
		tx.First(&g, gid)
		tx.Model(&g).Update("owner_id", trade.ReceiverID)
		proposerSpaceChange -= g.Salary
		receiverSpaceChange += g.Salary
	}

	// Transfer generals from receiver to proposer
	for _, gid := range requestGenerals {
		var g model.General
		tx.First(&g, gid)
		tx.Model(&g).Update("owner_id", trade.ProposerID)
		receiverSpaceChange -= g.Salary
		proposerSpaceChange += g.Salary
	}

	// Transfer treasures from proposer to receiver
	for _, tid := range offerTreasures {
		tx.Model(&model.Treasure{}).Where("id = ?", tid).Update("owner_id", trade.ReceiverID)
	}

	// Transfer treasures from receiver to proposer
	for _, tid := range requestTreasures {
		tx.Model(&model.Treasure{}).Where("id = ?", tid).Update("owner_id", trade.ProposerID)
	}

	// Handle space exchange
	proposerSpaceChange += trade.RequestSpace - trade.OfferSpace
	receiverSpaceChange += trade.OfferSpace - trade.RequestSpace

	// Update used space
	tx.Model(&proposer).Update("used_space", proposer.UsedSpace+proposerSpaceChange)
	tx.Model(&receiver).Update("used_space", receiver.UsedSpace+receiverSpaceChange)

	// Update trade status
	tx.Model(&trade).Update("status", "accepted")

	if err := tx.Commit().Error; err != nil {
		return err
	}

	// Log the trade acceptance
	logTrade(tradeID, "accepted", userID, "Trade accepted")

	return nil
}

// RejectTrade rejects a trade proposal
func RejectTrade(tradeID uint, userID uint) error {
	db := database.GetDB()

	var trade model.Trade
	if err := db.First(&trade, tradeID).Error; err != nil {
		return ErrTradeNotFound
	}

	// Check if user is the receiver
	if trade.ReceiverID != userID {
		return ErrNotTradeParticipant
	}

	// Check if trade is still pending
	if trade.Status != "pending" {
		return ErrTradeAlreadyProcessed
	}

	// Update trade status
	if err := db.Model(&trade).Update("status", "rejected").Error; err != nil {
		return err
	}

	// Log the trade rejection
	logTrade(tradeID, "rejected", userID, "Trade rejected")

	return nil
}

// CancelTrade cancels a trade proposal by proposer
func CancelTrade(tradeID uint, userID uint) error {
	db := database.GetDB()

	var trade model.Trade
	if err := db.First(&trade, tradeID).Error; err != nil {
		return ErrTradeNotFound
	}

	// Check if user is the proposer
	if trade.ProposerID != userID {
		return ErrNotTradeParticipant
	}

	// Check if trade is still pending
	if trade.Status != "pending" {
		return ErrTradeAlreadyProcessed
	}

	// Update trade status
	if err := db.Model(&trade).Update("status", "cancelled").Error; err != nil {
		return err
	}

	// Log the trade cancellation
	logTrade(tradeID, "cancelled", userID, "Trade cancelled")

	return nil
}

// GetPendingTrades returns pending trades for a user
func GetPendingTrades(userID uint) ([]model.Trade, error) {
	db := database.GetDB()

	var trades []model.Trade
	if err := db.Where("(proposer_id = ? OR receiver_id = ?) AND status = ?", userID, userID, "pending").
		Preload("Proposer").Preload("Receiver").
		Order("created_at DESC").
		Find(&trades).Error; err != nil {
		return nil, err
	}

	return trades, nil
}

// GetTradeHistory returns trade history for a user
func GetTradeHistory(userID uint) ([]model.Trade, error) {
	db := database.GetDB()

	var trades []model.Trade
	if err := db.Where("proposer_id = ? OR receiver_id = ?", userID, userID).
		Preload("Proposer").Preload("Receiver").
		Order("created_at DESC").
		Find(&trades).Error; err != nil {
		return nil, err
	}

	return trades, nil
}

// GetAllTrades returns all trades (admin)
func GetAllTrades() ([]model.Trade, error) {
	db := database.GetDB()

	var trades []model.Trade
	if err := db.Preload("Proposer").Preload("Receiver").
		Order("created_at DESC").
		Find(&trades).Error; err != nil {
		return nil, err
	}

	return trades, nil
}

// GetTradeByID returns a trade by ID
func GetTradeByID(tradeID uint) (*model.Trade, error) {
	db := database.GetDB()

	var trade model.Trade
	if err := db.Preload("Proposer").Preload("Receiver").First(&trade, tradeID).Error; err != nil {
		return nil, ErrTradeNotFound
	}

	return &trade, nil
}

// logTrade logs a trade action
func logTrade(tradeID uint, action string, performedBy uint, details string) {
	db := database.GetDB()

	log := model.TradeLog{
		TradeID:     tradeID,
		Action:      action,
		PerformedBy: performedBy,
		Details:     details,
	}

	db.Create(&log)
}
