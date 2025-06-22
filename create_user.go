package main

import (
	"log"

	"github.com/kart2405/API_Gateway/internal/config"
	"github.com/kart2405/API_Gateway/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Initialize database
	config.InitDatabase()

	// Create a test user
	username := "test"
	password := "test"

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Check if user already exists
	var existingUser models.User
	if err := config.DB.Where("username = ?", username).First(&existingUser).Error; err == nil {
		log.Printf("User '%s' already exists", username)
		return
	}

	// Create new user
	user := models.User{
		Username: username,
		Password: string(hashedPassword),
	}

	if err := config.DB.Create(&user).Error; err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}

	log.Printf("User '%s' created successfully with ID: %d", username, user.ID)
}
