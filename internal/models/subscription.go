package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SubscriptionPlan represents subscription plans associated with products
type SubscriptionPlan struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProductID uuid.UUID  `gorm:"type:uuid;not null;index" json:"product_id" validate:"required"`
	PlanName  string     `gorm:"not null;size:255" json:"plan_name" validate:"required,min=1,max=255"`
	Duration  int32      `gorm:"not null" json:"duration" validate:"required,min=1"` // in days
	Price     float64    `gorm:"not null;type:decimal(10,2)" json:"price" validate:"required,min=0"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	
	// Relationship
	Product Product `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE;" json:"product,omitempty"`
}

func (SubscriptionPlan) TableName() string {
	return "subscription_plan"
}

func (s *SubscriptionPlan) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
