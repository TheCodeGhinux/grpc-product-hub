// Package types contains type definition for product and subscription plan management.
package types

import (
	"context"
	"grpc-product/internal/models"
	"grpc-product/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type ProductService interface {
	CreateProduct(ctx context.Context, req *CreateProductRequest) (*models.Product, error)
	GetProduct(ctx context.Context, id uuid.UUID) (*models.Product, error)
	UpdateProduct(ctx context.Context, id uuid.UUID, req *UpdateProductRequest) (*models.Product, error)
	DeleteProduct(ctx context.Context, id uuid.UUID) error
	ListProducts(ctx context.Context, filters *ListProductsRequest) (*ListProductsResponse, error)
	ValidateProduct(product *models.Product) error
}

type SubscriptionPlanService interface {
	CreateSubscriptionPlan(ctx context.Context, req *CreateSubscriptionPlanRequest) (*models.SubscriptionPlan, error)
	GetSubscriptionPlan(ctx context.Context, id uuid.UUID) (*models.SubscriptionPlan, error)
	GetProductSubscriptionPlans(ctx context.Context, productID uuid.UUID) ([]*models.SubscriptionPlan, error)
	UpdateSubscriptionPlan(ctx context.Context, id uuid.UUID, req *UpdateSubscriptionPlanRequest) (*models.SubscriptionPlan, error)
	DeleteSubscriptionPlan(ctx context.Context, id uuid.UUID) error
	ListSubscriptionPlans(ctx context.Context, filters *ListSubscriptionPlansRequest) (*ListSubscriptionPlansResponse, error)
}

type ListSubscriptionPlansRequest struct {
	ProductID   *uuid.UUID `json:"product_id,omitempty"`
	MinDuration *int32     `json:"min_duration,omitempty" validate:"omitempty,min=1"`
	MaxDuration *int32     `json:"max_duration,omitempty" validate:"omitempty,min=1"`
	MinPrice    *float64   `json:"min_price,omitempty" validate:"omitempty,min=0"`
	MaxPrice    *float64   `json:"max_price,omitempty" validate:"omitempty,min=0"`
	Page        int        `json:"page" validate:"min=1"`
	PageSize    int        `json:"page_size" validate:"min=1,max=100"`
	SortBy      string     `json:"sort_by" validate:"omitempty,oneof=plan_name duration price created_at"`
	SortDesc    bool       `json:"sort_desc"`
}

type UpdateSubscriptionPlanRequest struct {
	PlanName string  `json:"plan_name" validate:"required,min=1,max=255"`
	Duration int32   `json:"duration" validate:"required,min=1"`
	Price    float64 `json:"price" validate:"required,min=0"`
}

type CreateProductRequest struct {
	Name                string                             `json:"name" validate:"required,min=1,max=255"`
	Description         string                             `json:"description" validate:"max=1000"`
	Price               float64                            `json:"price" validate:"required,min=0"`
	Type                models.ProductType                 `json:"type" validate:"required,oneof=digital physical subscription"`
	DigitalDetails      *models.DigitalProductDetails      `json:"digital_details,omitempty"`
	PhysicalDetails     *models.PhysicalProductDetails     `json:"physical_details,omitempty"`
	SubscriptionDetails *models.SubscriptionProductDetails `json:"subscription_details,omitempty"`
}

type UpdateProductRequest struct {
	Name                string                             `json:"name" validate:"required,min=1,max=255"`
	Description         string                             `json:"description" validate:"omitempty,max=1000"`
	Price               float64                            `json:"price" validate:"omitempty,min=0"`
	DigitalDetails      *models.DigitalProductDetails      `json:"digital_details,omitempty"`
	PhysicalDetails     *models.PhysicalProductDetails     `json:"physical_details,omitempty"`
	SubscriptionDetails *models.SubscriptionProductDetails `json:"subscription_details,omitempty"`
}

type ListProductsRequest struct {
	Type     *models.ProductType `json:"type,omitempty"`
	Page     int                 `json:"page" validate:"min=1"`
	PageSize int                 `json:"page_size" validate:"min=1,max=100"`
	SortBy   string              `json:"sort_by" validate:"omitempty,oneof=name price created_at updated_at"`
	SortDesc bool                `json:"sort_desc"`
}

type ListProductsResponse struct {
	Products   []*models.Product `json:"products"`
	TotalCount int64             `json:"total_count"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

type CreateSubscriptionPlanRequest struct {
	ProductID uuid.UUID `json:"product_id" validate:"required"`
	PlanName  string    `json:"plan_name" validate:"required,min=1,max=255"`
	Duration  int32     `json:"duration" validate:"required,min=1"`
	Price     float64   `json:"price" validate:"required,min=0"`
}

type ListSubscriptionPlansResponse struct {
	Plans      []*models.SubscriptionPlan `json:"plans"`
	TotalCount int64                      `json:"total_count"`
	Page       int                        `json:"page"`
	PageSize   int                        `json:"page_size"`
	TotalPages int                        `json:"total_pages"`
}

type productService struct {
	productRepo repository.ProductRepository
	validator   *validator.Validate
}

type subscriptionPlanService struct {
	subscriptionRepo repository.SubscriptionPlanRepository
	productRepo      repository.ProductRepository
	validator        *validator.Validate
}
