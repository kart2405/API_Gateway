package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	viper.SetConfigFile("config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		viper.GetString("postgres.host"),
		viper.GetString("postgres.port"),
		viper.GetString("postgres.user"),
		viper.GetString("postgres.password"),
		viper.GetString("postgres.dbname"),
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	DB = db
	log.Println("Connected to database")
}
