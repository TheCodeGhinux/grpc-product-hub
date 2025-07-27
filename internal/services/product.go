// Package service contains business logic implementations for product and subscription plan management.
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

type productService struct {
	productRepo repository.ProductRepository
	validator   *validator.Validate
}


func NewProductService(productRepo repository.ProductRepository) types.ProductService {
	return &productService{
		productRepo: productRepo,
		validator:   validator.New(),
	}
}


func (s *productService) CreateProduct(ctx context.Context, req *types.CreateProductRequest) (*models.Product, error) {
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	
	product := &models.Product{
		ID:          uuid.New(), 
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Type:        req.Type,
	}

	
	switch req.Type {
	case models.ProductTypeDigital:
		if req.DigitalDetails == nil {
			return nil, errors.New("digital product details are required for digital products")
		}
		if err := s.validator.Struct(req.DigitalDetails); err != nil {
			return nil, fmt.Errorf("digital details validation failed: %w", err)
		}
		
		req.DigitalDetails.ID = uuid.New()
		req.DigitalDetails.ProductID = product.ID
		product.DigitalDetails = req.DigitalDetails

	case models.ProductTypePhysical:
		if req.PhysicalDetails == nil {
			return nil, errors.New("physical product details are required for physical products")
		}
		if err := s.validator.Struct(req.PhysicalDetails); err != nil {
			return nil, fmt.Errorf("physical details validation failed: %w", err)
		}
		
		req.PhysicalDetails.ID = uuid.New()
		req.PhysicalDetails.ProductID = product.ID
		product.PhysicalDetails = req.PhysicalDetails

	case models.ProductTypeSubscription:
		if req.SubscriptionDetails == nil {
			return nil, errors.New("subscription product details are required for subscription products")
		}
		if err := s.validator.Struct(req.SubscriptionDetails); err != nil {
			return nil, fmt.Errorf("subscription details validation failed: %w", err)
		}
		
		req.SubscriptionDetails.ID = uuid.New()
		req.SubscriptionDetails.ProductID = product.ID
		product.SubscriptionDetails = req.SubscriptionDetails

	default:
		return nil, fmt.Errorf("unsupported product type: %s", req.Type)
	}

	
	if err := product.Validate(); err != nil {
		return nil, fmt.Errorf("product validation failed: %w", err)
	}

	
	if err := s.productRepo.Create(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return product, nil
}

func (s *productService) GetProduct(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	if id == uuid.Nil {
		return nil, errors.New("invalid product ID")
	}

	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return product, nil
}


func (s *productService) UpdateProduct(ctx context.Context, id uuid.UUID, req *types.UpdateProductRequest) (*models.Product, error) {
	if id == uuid.Nil {
		return nil, errors.New("invalid product ID")
	}

	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	
	existingProduct, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing product: %w", err)
	}

	
	existingProduct.Name = req.Name
	existingProduct.Description = req.Description
	existingProduct.Price = req.Price

	
	switch existingProduct.Type {
	case models.ProductTypeDigital:
		if req.DigitalDetails != nil {
			if err := s.validator.Struct(req.DigitalDetails); err != nil {
				return nil, fmt.Errorf("digital details validation failed: %w", err)
			}
			existingProduct.DigitalDetails = req.DigitalDetails
		}

	case models.ProductTypePhysical:
		if req.PhysicalDetails != nil {
			if err := s.validator.Struct(req.PhysicalDetails); err != nil {
				return nil, fmt.Errorf("physical details validation failed: %w", err)
			}
			existingProduct.PhysicalDetails = req.PhysicalDetails
		}

	case models.ProductTypeSubscription:
		if req.SubscriptionDetails != nil {
			if err := s.validator.Struct(req.SubscriptionDetails); err != nil {
				return nil, fmt.Errorf("subscription details validation failed: %w", err)
			}
			existingProduct.SubscriptionDetails = req.SubscriptionDetails
		}
	}

	
	if err := s.ValidateProduct(existingProduct); err != nil {
		return nil, fmt.Errorf("business validation failed: %w", err)
	}

	
	if err := s.productRepo.Update(ctx, existingProduct); err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return existingProduct, nil
}

func (s *productService) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("invalid product ID")
	}

	
	_, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("product not found: %w", err)
	}

	
	if err := s.productRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}


func (s *productService) ListProducts(ctx context.Context, req *types.ListProductsRequest) (*types.ListProductsResponse, error) {
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	
	filters := repository.ProductFilters{
		Type:     req.Type,
		Page:     req.Page,
		PageSize: req.PageSize,
		SortBy:   req.SortBy,
		SortDesc: req.SortDesc,
	}

	
	products, totalCount, err := s.productRepo.List(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	
	totalPages := int(totalCount) / req.PageSize
	if int(totalCount)%req.PageSize > 0 {
		totalPages++
	}

	return &types.ListProductsResponse{
		Products:   products,
		TotalCount: totalCount,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}


func (s *productService) ValidateProduct(product *models.Product) error {
	if product == nil {
		return errors.New("product cannot be nil")
	}

	
	if product.Price < 0 {
		return errors.New("product price cannot be negative")
	}

	
	switch product.Type {
	case models.ProductTypeDigital:
		if product.DigitalDetails == nil {
			return errors.New("digital products must have digital details")
		}
		if product.DigitalDetails.FileSize <= 0 {
			return errors.New("digital product file size must be positive")
		}

	case models.ProductTypePhysical:
		if product.PhysicalDetails == nil {
			return errors.New("physical products must have physical details")
		}
		if product.PhysicalDetails.Weight <= 0 {
			return errors.New("physical product weight must be positive")
		}

	case models.ProductTypeSubscription:
		if product.SubscriptionDetails == nil {
			return errors.New("subscription products must have subscription details")
		}
		if product.SubscriptionDetails.RenewalPrice < 0 {
			return errors.New("subscription renewal price cannot be negative")
		}
		
		validPeriods := []string{"daily", "weekly", "monthly", "quarterly", "yearly"}
		isValidPeriod := false
		for _, period := range validPeriods {
			if product.SubscriptionDetails.SubscriptionPeriod == period {
				isValidPeriod = true
				break
			}
		}
		if !isValidPeriod {
			return errors.New("invalid subscription period")
		}
	}

	return nil
}
