package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Model struct {
	ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt"`
}
type User struct {
	Model
	Email    string `gorm:"unique;not null" json:"email"`
	Password string `gorm:"not null;check:length(password)=60" json:"-"`
}
type FileMetadata struct {
	Model
	UserID     uuid.UUID `gorm:"type:uuid;not null"`
	User       User      `gorm:"foreignKey:UserID"`
	Filename   string    `gorm:"not null"`
	S3Key      string    `gorm:"not null;uniqueIndex"` // UserID/Filename
	IsPassword bool      `gorm:"not null;default:false"`
	Password   string
	ExpiryTime time.Time `gorm:"not null"`
}
