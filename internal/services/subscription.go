package service

import (
	"context"
	"errors"
	"fmt"
	"grpc-product/internal/models"
	"grpc-product/internal/repository"
	"grpc-product/internal/types"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// subscriptionPlanService implements types.SubscriptionPlanService
type subscriptionPlanService struct {
	subscriptionRepo      repository.SubscriptionPlanRepository
	validator *validator.Validate
}

// NewSubscriptionPlanService creates a new SubscriptionPlanService
func NewSubscriptionPlanService(subscriptionRepo repository.SubscriptionPlanRepository) types.SubscriptionPlanService {
	return &subscriptionPlanService{
		subscriptionRepo: subscriptionRepo,
		validator: validator.New(),
	}
}

// CreateSubscriptionPlan creates a new subscription plan
func (s *subscriptionPlanService) CreateSubscriptionPlan(ctx context.Context, req *types.CreateSubscriptionPlanRequest) (*models.SubscriptionPlan, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Build model
	plan := &models.SubscriptionPlan{
		ID:        uuid.New(),
		ProductID: req.ProductID,
		PlanName:  req.PlanName,
		Duration:  req.Duration,
		Price:     req.Price,
	}

	// Save
	if err := s.subscriptionRepo.Create(ctx, plan); err != nil {
		return nil, fmt.Errorf("failed to create subscription plan: %w", err)
	}

	return plan, nil
}

// GetSubscriptionPlan retrieves a plan by ID
func (s *subscriptionPlanService) GetSubscriptionPlan(ctx context.Context, id uuid.UUID) (*models.SubscriptionPlan, error) {
	if id == uuid.Nil {
		return nil, errors.New("invalid subscription plan ID")
	}

	plan, err := s.subscriptionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription plan: %w", err)
	}

	return plan, nil
}

// UpdateSubscriptionPlan updates an existing subscription plan
func (s *subscriptionPlanService) UpdateSubscriptionPlan(ctx context.Context, id uuid.UUID, req *types.UpdateSubscriptionPlanRequest) (*models.SubscriptionPlan, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if id == uuid.Nil {
		return nil, errors.New("invalid subscription plan ID")
	}

	// Fetch existing
	existing, err := s.subscriptionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing plan: %w", err)
	}

	// Apply updates
	existing.PlanName = req.PlanName
	existing.Duration = req.Duration
	existing.Price = req.Price

	// Validate business rules (optionally add here)
	// For example ensure duration > 0 and price >= 0
	if existing.Duration <= 0 {
		return nil, errors.New("duration must be positive")
	}
	if existing.Price < 0 {
		return nil, errors.New("price cannot be negative")
	}

	// Save
	if err := s.subscriptionRepo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("failed to update subscription plan: %w", err)
	}

	return existing, nil
}

// DeleteSubscriptionPlan soft-deletes a plan
func (s *subscriptionPlanService) DeleteSubscriptionPlan(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("invalid subscription plan ID")
	}

	// ensure exists
	if _, err := s.subscriptionRepo.GetByID(ctx, id); err != nil {
		return fmt.Errorf("subscription plan not found: %w", err)
	}

	if err := s.subscriptionRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete subscription plan: %w", err)
	}

	return nil
}

// ListSubscriptionPlans returns paginated plans
func (s *subscriptionPlanService) ListSubscriptionPlans(ctx context.Context, req *types.ListSubscriptionPlansRequest) (*types.ListSubscriptionPlansResponse, error) {
	// Validate
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// filters
	filters := repository.SubscriptionPlanFilters{
		Page:     req.Page,
		PageSize: req.PageSize,
		SortBy:   req.SortBy,
		SortDesc: req.SortDesc,
	}
	if req.ProductID != nil {
		filters.ProductID = req.ProductID
	}

	plans, total, err := s.subscriptionRepo.List(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscription plans: %w", err)
	}

	// compute pages
	totalPages := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPages++
	}

	return &types.ListSubscriptionPlansResponse{
		Plans:      plans,
		TotalCount: total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetProductSubscriptionPlans retrieves plans by product
func (s *subscriptionPlanService) GetProductSubscriptionPlans(ctx context.Context, productID uuid.UUID) ([]*models.SubscriptionPlan, error) {
	if productID == uuid.Nil {
		return nil, errors.New("invalid product ID")
	}

	plans, err := s.subscriptionRepo.GetByProductID(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plans by product ID: %w", err)
	}
	return plans, nil
}
