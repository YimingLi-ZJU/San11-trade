package config

import (
	"os"
)

// Config holds the application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Game     GameConfig
}

type ServerConfig struct {
	Port string
	Host string
}

type DatabaseConfig struct {
	Path string
}

type JWTConfig struct {
	Secret     string
	ExpireHour int
}

type GameConfig struct {
	InitialSpace     int // Initial space for each player (default 350)
	GuaranteeDraws   int // Number of guarantee draws (default 3)
	NormalDraws      int // Number of normal draws (default 7)
	DraftRounds      int // Number of draft rounds (default 4)
	PlayersPerSeason int // Number of players per season (default 32)
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},
		Database: DatabaseConfig{
			Path: getEnv("DB_PATH", "./data/san11trade.db"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "san11-trade-secret-key-change-in-production"),
			ExpireHour: 72,
		},
		Game: GameConfig{
			InitialSpace:     350,
			GuaranteeDraws:   3,
			NormalDraws:      7,
			DraftRounds:      4,
			PlayersPerSeason: 32,
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Global config instance
var AppConfig *Config

func Init() {
	AppConfig = DefaultConfig()
}
