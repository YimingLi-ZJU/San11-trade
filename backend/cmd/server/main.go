package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"san11-trade/internal/api"
	"san11-trade/internal/config"
	"san11-trade/internal/database"
	"san11-trade/internal/service"
)

func main() {
	// Command line flags
	var (
		port        = flag.String("port", "", "Server port (default: 8080)")
		dbPath      = flag.String("db", "", "Database path (default: ./data/san11trade.db)")
		createAdmin = flag.Bool("create-admin", false, "Create admin user")
		adminUser   = flag.String("admin-user", "admin", "Admin username")
		adminPass   = flag.String("admin-pass", "admin123", "Admin password")
	)
	flag.Parse()

	// Initialize configuration
	config.Init()

	// Override config with command line flags
	if *port != "" {
		config.AppConfig.Server.Port = *port
	}
	if *dbPath != "" {
		config.AppConfig.Database.Path = *dbPath
	}

	// Initialize database
	if err := database.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create admin user if requested
	if *createAdmin {
		if err := service.CreateAdmin(*adminUser, *adminPass); err != nil {
			log.Printf("Failed to create admin: %v", err)
		} else {
			log.Printf("Admin user '%s' created/updated successfully", *adminUser)
		}
		if !*createAdmin {
			os.Exit(0)
		}
	}

	// Setup router
	router := api.SetupRouter()

	// Start server
	addr := fmt.Sprintf("%s:%s", config.AppConfig.Server.Host, config.AppConfig.Server.Port)
	log.Printf("Starting server on %s", addr)
	log.Printf("API documentation: http://%s/health", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
