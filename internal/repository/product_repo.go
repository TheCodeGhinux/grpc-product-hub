// Package repository provides data access implementations for product and subscription plan entities.
package repository

import (
	"context"
	"errors"
	"fmt"
	"grpc-product/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)


type ProductRepository interface {
	Create(ctx context.Context, product *models.Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error)
	Update(ctx context.Context, product *models.Product) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filters ProductFilters) ([]*models.Product, int64, error)
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
}


type ProductFilters struct {
	Type     *models.ProductType
	Page     int
	PageSize int
	SortBy   string
	SortDesc bool
}


type productRepository struct {
	db *gorm.DB
}


func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{
		db: db,
	}
}


func (r *productRepository) Create(ctx context.Context, product *models.Product) error {
	if product == nil {
		return errors.New("product cannot be nil")
	}

	
	if err := product.Validate(); err != nil {
		return fmt.Errorf("product validation failed: %w", err)
	}

	
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		
		
		if err := tx.Create(product).Error; err != nil {
			return fmt.Errorf("failed to create product: %w", err)
		}

		
		
		
		return nil
	})
}


func (r *productRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	if id == uuid.Nil {
		return nil, errors.New("invalid product ID")
	}

	var product models.Product

	query := r.db.WithContext(ctx).
		Preload("DigitalDetails").
		Preload("PhysicalDetails").
		Preload("SubscriptionDetails").
		Preload("SubscriptionPlans")

	if err := query.First(&product, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("product with ID %s not found", id.String())
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &product, nil
}


func (r *productRepository) Update(ctx context.Context, product *models.Product) error {
	if product == nil {
		return errors.New("product cannot be nil")
	}

	if product.ID == uuid.Nil {
		return errors.New("product ID is required for update")
	}

	
	if err := product.Validate(); err != nil {
		return fmt.Errorf("product validation failed: %w", err)
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		
		var existingProduct models.Product
		if err := tx.Preload("DigitalDetails").Preload("PhysicalDetails").Preload("SubscriptionDetails").First(&existingProduct, "id = ?", product.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("product with ID %s not found", product.ID.String())
			}
			return fmt.Errorf("failed to check product existence: %w", err)
		}

		
		
		if existingProduct.Type != product.Type {
			if err := r.cleanupOldDetails(tx, existingProduct.ID, existingProduct.Type); err != nil {
				return fmt.Errorf("failed to cleanup old details: %w", err)
			}
		}

		
		updates := map[string]interface{}{
			"name":        product.Name,
			"description": product.Description,
			"price":       product.Price,
			"type":        product.Type,
		}

		if err := tx.Model(&existingProduct).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update product: %w", err)
		}

		
		switch product.Type {
		case models.ProductTypeDigital:
			if product.DigitalDetails != nil {
				product.DigitalDetails.ProductID = product.ID
				if existingProduct.Type == product.Type && existingProduct.DigitalDetails != nil {
					
					product.DigitalDetails.ID = existingProduct.DigitalDetails.ID
				}
				if err := tx.Save(product.DigitalDetails).Error; err != nil {
					return fmt.Errorf("failed to update digital product details: %w", err)
				}
			}
		case models.ProductTypePhysical:
			if product.PhysicalDetails != nil {
				product.PhysicalDetails.ProductID = product.ID
				if existingProduct.Type == product.Type && existingProduct.PhysicalDetails != nil {
					
					product.PhysicalDetails.ID = existingProduct.PhysicalDetails.ID
				}
				if err := tx.Save(product.PhysicalDetails).Error; err != nil {
					return fmt.Errorf("failed to update physical product details: %w", err)
				}
			}
		case models.ProductTypeSubscription:
			if product.SubscriptionDetails != nil {
				product.SubscriptionDetails.ProductID = product.ID
				if existingProduct.Type == product.Type && existingProduct.SubscriptionDetails != nil {
					
					product.SubscriptionDetails.ID = existingProduct.SubscriptionDetails.ID
				}
				if err := tx.Save(product.SubscriptionDetails).Error; err != nil {
					return fmt.Errorf("failed to update subscription product details: %w", err)
				}
			}
		}

		return nil
	})
}


func (r *productRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("invalid product ID")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		
		var product models.Product
		if err := tx.First(&product, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("product with ID %s not found", id.String())
			}
			return fmt.Errorf("failed to check product existence: %w", err)
		}

		
		if err := tx.Delete(&product).Error; err != nil {
			return fmt.Errorf("failed to delete product: %w", err)
		}

		return nil
	})
}


func (r *productRepository) List(ctx context.Context, filters ProductFilters) ([]*models.Product, int64, error) {
	var products []*models.Product
	var totalCount int64

	
	query := r.db.WithContext(ctx).Model(&models.Product{})

	
	if filters.Type != nil {
		query = query.Where("type = ?", *filters.Type)
	}

	
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
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

	
	if err := query.
		Preload("DigitalDetails").
		Preload("PhysicalDetails").
		Preload("SubscriptionDetails").
		Preload("SubscriptionPlans").
		Find(&products).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}

	return products, totalCount, nil
}


func (r *productRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	if id == uuid.Nil {
		return false, errors.New("invalid product ID")
	}

	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Product{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check product existence: %w", err)
	}

	return count > 0, nil
}


func (r *productRepository) cleanupOldDetails(tx *gorm.DB, productID uuid.UUID, oldType models.ProductType) error {
	switch oldType {
	case models.ProductTypeDigital:
		return tx.Where("product_id = ?", productID).Delete(&models.DigitalProductDetails{}).Error
	case models.ProductTypePhysical:
		return tx.Where("product_id = ?", productID).Delete(&models.PhysicalProductDetails{}).Error
	case models.ProductTypeSubscription:
		return tx.Where("product_id = ?", productID).Delete(&models.SubscriptionProductDetails{}).Error
	}
	return nil
}