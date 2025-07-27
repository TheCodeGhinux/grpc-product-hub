package repository

import (
	"context"
	"errors"
	"fmt"
	"grpc-product/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SubscriptionPlanRepository interface {
	Create(ctx context.Context, plan *models.SubscriptionPlan) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.SubscriptionPlan, error)
	Update(ctx context.Context, plan *models.SubscriptionPlan) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filters SubscriptionPlanFilters) ([]*models.SubscriptionPlan, int64, error)
	GetByProductID(ctx context.Context, productID uuid.UUID) ([]*models.SubscriptionPlan, error)
}

type SubscriptionPlanFilters struct {
	ProductID *uuid.UUID
	Page      int
	PageSize  int
	SortBy    string
	SortDesc  bool
}

type subscriptionPlanRepository struct {
	db *gorm.DB
}

func NewSubscriptionPlanRepository(db *gorm.DB) SubscriptionPlanRepository {
	return &subscriptionPlanRepository{db: db}
}

func (r *subscriptionPlanRepository) Create(ctx context.Context, plan *models.SubscriptionPlan) error {
	if plan == nil {
		return errors.New("subscription plan cannot be nil")
	}
	return r.db.WithContext(ctx).Create(plan).Error
}

func (r *subscriptionPlanRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.SubscriptionPlan, error) {
	if id == uuid.Nil {
		return nil, errors.New("invalid plan ID")
	}

	var plan models.SubscriptionPlan
	err := r.db.WithContext(ctx).Preload("Product").First(&plan, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("subscription plan with ID %s not found", id.String())
		}
		return nil, fmt.Errorf("failed to get subscription plan: %w", err)
	}

	return &plan, nil
}

func (r *subscriptionPlanRepository) Update(ctx context.Context, plan *models.SubscriptionPlan) error {
	if plan == nil {
		return errors.New("subscription plan cannot be nil")
	}
	if plan.ID == uuid.Nil {
		return errors.New("subscription plan ID is required for update")
	}

	return r.db.WithContext(ctx).Model(plan).Updates(map[string]interface{}{
		"plan_name": plan.PlanName,
		"duration":  plan.Duration,
		"price":     plan.Price,
	}).Error
}

func (r *subscriptionPlanRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("invalid plan ID")
	}

	return r.db.WithContext(ctx).Delete(&models.SubscriptionPlan{}, "id = ?", id).Error
}

func (r *subscriptionPlanRepository) List(ctx context.Context, filters SubscriptionPlanFilters) ([]*models.SubscriptionPlan, int64, error) {
	var plans []*models.SubscriptionPlan
	var total int64

	query := r.db.WithContext(ctx).Model(&models.SubscriptionPlan{})

	if filters.ProductID != nil {
		query = query.Where("product_id = ?", *filters.ProductID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count subscription plans: %w", err)
	}

	sortBy := "created_at"
	if filters.SortBy != "" {
		sortBy = filters.SortBy
	}
	sortOrder := "ASC"
	if filters.SortDesc {
		sortOrder = "DESC"
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	if filters.PageSize > 0 {
		offset := 0
		if filters.Page > 0 {
			offset = (filters.Page - 1) * filters.PageSize
		}
		query = query.Offset(offset).Limit(filters.PageSize)
	}

	if err := query.Preload("Product").Find(&plans).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list subscription plans: %w", err)
	}

	return plans, total, nil
}

func (r *subscriptionPlanRepository) GetByProductID(ctx context.Context, productID uuid.UUID) ([]*models.SubscriptionPlan, error) {
	if productID == uuid.Nil {
		return nil, errors.New("invalid product ID")
	}

	var plans []*models.SubscriptionPlan
	err := r.db.WithContext(ctx).
		Where("product_id = ?", productID).
		Order("created_at DESC").
		Find(&plans).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get plans for product: %w", err)
	}

	return plans, nil
}
