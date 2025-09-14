package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserProfile struct {
	ID          string         `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	UserID      string         `json:"user_id" gorm:"type:uuid;not null;unique"`
	User        User           `json:"user" gorm:"foreignKey:UserID"`
	FirstName   string         `json:"first_name"`
	LastName    string         `json:"last_name"`
	Phone       string         `json:"phone"`
	Avatar      string         `json:"avatar"`
	DateOfBirth *time.Time     `json:"date_of_birth"`
	Gender      string         `json:"gender" gorm:"type:varchar(10)"`
	Address     string         `json:"address"`
	City        string         `json:"city"`
	State       string         `json:"state"`
	Country     string         `json:"country"`
	ZipCode     string         `json:"zip_code"`
	Bio         string         `json:"bio"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

func (UserProfile) TableName() string {
	return "user_profiles"
}

func (up *UserProfile) BeforeCreate(tx *gorm.DB) error {
	if up.ID == "" {
		up.ID = uuid.NewString()
	}
	return nil
}
