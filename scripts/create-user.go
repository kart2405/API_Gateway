package main

import (
	"log"

	"github.com/kart2405/API_Gateway/internal/config"
	"github.com/kart2405/API_Gateway/internal/models"
	"github.com/kart2405/API_Gateway/internal/services"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Initialize database
	config.InitDatabase()
	config.DB.AutoMigrate(&models.User{}, &services.RouteConfig{})

	// Check if admin user already exists
	var existingUser models.User
	if err := config.DB.Where("username = ?", "admin").First(&existingUser).Error; err == nil {
		log.Println("Admin user already exists")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Create admin user
	adminUser := models.User{
		Username: "admin",
		Password: string(hashedPassword),
	}

	if err := config.DB.Create(&adminUser).Error; err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	log.Println("Admin user created successfully")
	log.Println("Username: admin")
	log.Println("Password: admin123")
}
