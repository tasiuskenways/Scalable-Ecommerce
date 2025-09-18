package entities

import (
	"time"

	"gorm.io/gorm"
)

type InvitationStatus string

const (
	InvitationStatusPending  InvitationStatus = "PENDING"
	InvitationStatusAccepted InvitationStatus = "ACCEPTED"
	InvitationStatusDeclined InvitationStatus = "DECLINED"
	InvitationStatusExpired  InvitationStatus = "EXPIRED"
)

type StoreInvitation struct {
	ID         string           `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	StoreID    string           `json:"store_id" gorm:"not null;index"`
	InviterID  string           `json:"inviter_id" gorm:"not null;index"`
	InviteeID  *string          `json:"invitee_id,omitempty" gorm:"index"` // Nullable for email invitations
	Email      string           `json:"email" gorm:"not null;size:100"`
	Role       StoreRole        `json:"role" gorm:"not null;type:varchar(20)"`
	Status     InvitationStatus `json:"status" gorm:"not null;type:varchar(20);default:'PENDING'"`
	Token      string           `json:"-" gorm:"not null;uniqueIndex;size:64"` // Hidden from JSON
	ExpiresAt  time.Time        `json:"expires_at" gorm:"not null"`
	AcceptedAt *time.Time       `json:"accepted_at,omitempty"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
	DeletedAt  gorm.DeletedAt   `json:"-" gorm:"index"`

	// Relationships
	Store *Store `json:"store,omitempty" gorm:"foreignKey:StoreID"`
}

func (StoreInvitation) TableName() string {
	return "store_invitations"
}

// IsExpired checks if the invitation has expired
func (si *StoreInvitation) IsExpired() bool {
	return time.Now().After(si.ExpiresAt)
}

// CanAccept checks if the invitation can be accepted
func (si *StoreInvitation) CanAccept() bool {
	return si.Status == InvitationStatusPending && !si.IsExpired()
}