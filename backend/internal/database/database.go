package database

import (
	"log"
	"os"
	"path/filepath"

	"san11-trade/internal/config"
	"san11-trade/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Init initializes the database connection
func Init() error {
	cfg := config.AppConfig

	// Ensure the directory exists
	dbDir := filepath.Dir(cfg.Database.Path)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return err
	}

	var err error
	DB, err = gorm.Open(sqlite.Open(cfg.Database.Path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}

	log.Printf("Database connected: %s", cfg.Database.Path)

	// Auto migrate the schema
	if err := migrate(); err != nil {
		return err
	}

	// Initialize game phase if not exists
	if err := initGamePhase(); err != nil {
		return err
	}

	return nil
}

// migrate runs the database migrations
func migrate() error {
	return DB.AutoMigrate(
		&model.User{},
		&model.General{},
		&model.Treasure{},
		&model.City{},
		&model.Club{},
		&model.Policy{},
		&model.GameRule{},
		&model.Trade{},
		&model.GamePhase{},
		&model.DrawRecord{},
		&model.DraftRecord{},
		&model.TradeLog{},
	)
}

// initGamePhase creates the initial game phase record
func initGamePhase() error {
	var phase model.GamePhase
	result := DB.First(&phase)
	if result.Error == gorm.ErrRecordNotFound {
		phase = model.GamePhase{
			CurrentPhase: "signup",
			RoundNumber:  1,
			DraftRound:   0,
			DraftOrder:   "[]",
			Config:       "{}",
		}
		return DB.Create(&phase).Error
	}
	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}
