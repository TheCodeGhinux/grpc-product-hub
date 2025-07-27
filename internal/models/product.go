// Package models contains implementations for handling models
package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProductType string

const (
	ProductTypeDigital      ProductType = "digital"
	ProductTypePhysical     ProductType = "physical"
	ProductTypeSubscription ProductType = "subscription"
)

func (t *ProductType) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	switch s {
	case string(ProductTypeDigital), string(ProductTypePhysical), string(ProductTypeSubscription):
		*t = ProductType(s)
		return nil
	}

	return fmt.Errorf("invalid product type: %q", s)
}

type Product struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	Name        string     `gorm:"not null;size:255" json:"name" validate:"required,min=1,max=255"`
	Description string     `gorm:"type:text" json:"description"`
	Price       float64    `gorm:"not null;type:decimal(10,2)" json:"price" validate:"required,min=0"`
	Type        ProductType `gorm:"not null;type:varchar(20)" json:"type" validate:"required,oneof=digital physical subscription"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	
	DigitalDetails      *DigitalProductDetails      `gorm:"constraint:OnDelete:CASCADE;" json:"digital_details,omitempty"`
	PhysicalDetails     *PhysicalProductDetails     `gorm:"constraint:OnDelete:CASCADE;" json:"physical_details,omitempty"`
	SubscriptionDetails *SubscriptionProductDetails `gorm:"constraint:OnDelete:CASCADE;" json:"subscription_details,omitempty"`

	
	SubscriptionPlans []SubscriptionPlan `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE;" json:"subscription_plans,omitempty"`
}


type DigitalProductDetails struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	ProductID    uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
	FileSize     int64     `gorm:"not null" json:"file_size" validate:"required,min=1"`
	DownloadLink string    `gorm:"not null;size:500" json:"download_link" validate:"required,url,max=500"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}


type PhysicalProductDetails struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	ProductID  uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
	Weight     float64   `gorm:"not null;type:decimal(8,2)" json:"weight" validate:"required,min=0"`
	Dimensions string    `gorm:"not null;size:100" json:"dimensions" validate:"required,max=100"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}


type SubscriptionProductDetails struct {
	ID                 uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	ProductID          uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
	SubscriptionPeriod string    `gorm:"not null;size:50" json:"subscription_period" validate:"required,oneof=daily weekly monthly quarterly yearly"`
	RenewalPrice       float64   `gorm:"not null;type:decimal(10,2)" json:"renewal_price" validate:"required,min=0"`
	CreatedAt          time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Product) TableName() string {
	return "products"
}

func (DigitalProductDetails) TableName() string {
	return "digital_product_details"
}

func (PhysicalProductDetails) TableName() string {
	return "physical_product_details"
}

func (SubscriptionProductDetails) TableName() string {
	return "subscription_product_details"
}


func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

func (d *DigitalProductDetails) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

func (p *PhysicalProductDetails) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

func (s *SubscriptionProductDetails) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}


func (p *Product) IsDigital() bool {
	return p.Type == ProductTypeDigital
}

func (p *Product) IsPhysical() bool {
	return p.Type == ProductTypePhysical
}

func (p *Product) IsSubscription() bool {
	return p.Type == ProductTypeSubscription
}


func (p *Product) Validate() error {
	switch p.Type {
	case ProductTypeDigital:
		if p.DigitalDetails == nil {
			return gorm.ErrInvalidData
		}
		if p.PhysicalDetails != nil || p.SubscriptionDetails != nil {
			return gorm.ErrInvalidData
		}
	case ProductTypePhysical:
		if p.PhysicalDetails == nil {
			return gorm.ErrInvalidData
		}
		if p.DigitalDetails != nil || p.SubscriptionDetails != nil {
			return gorm.ErrInvalidData
		}
	case ProductTypeSubscription:
		if p.SubscriptionDetails == nil {
			return gorm.ErrInvalidData
		}
		if p.DigitalDetails != nil || p.PhysicalDetails != nil {
			return gorm.ErrInvalidData
		}
	}
	return nil
}