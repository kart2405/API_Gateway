package models

import "time"

type User struct {
	ID        uint   `gorm:"primaryKey"`
	Username  string `gorm:"uniqueIndex;not null"`
	Password  string `gorm:"not null"` // This will be a bcrypt hash
	CreatedAt time.Time
}
