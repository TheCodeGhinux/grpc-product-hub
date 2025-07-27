// Package database contains database configuration
package database

import (
	"fmt"
	"grpc-product/internal/models"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	DB *gorm.DB
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}


func NewDatabase(config DatabaseConfig) (*Database, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
	)

	
	var gormLogger logger.Interface
	if os.Getenv("ENV") == "development" {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Error)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	
	sqlDB.SetMaxIdleConns(10)

	
	sqlDB.SetMaxOpenConns(100)

	return &Database{DB: db}, nil
}


func GetDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "password"),
		DBName:   getEnv("DB_NAME", "grpc_product"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}


func (d *Database) AutoMigrate() error {
	log.Println("Running database migrations...")

	err := d.DB.AutoMigrate(
		&models.Product{},
		&models.DigitalProductDetails{},
		&models.PhysicalProductDetails{},
		&models.SubscriptionProductDetails{},
		&models.SubscriptionPlan{},
	)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}


func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}


func (d *Database) CreateIndexes() error {
	log.Println("Creating database indexes...")

	
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_products_type ON products(type);",
		"CREATE INDEX IF NOT EXISTS idx_products_created_at ON products(created_at);",
		"CREATE INDEX IF NOT EXISTS idx_subscription_plans_product_id ON subscription_plans(product_id);",
		"CREATE INDEX IF NOT EXISTS idx_subscription_plans_duration ON subscription_plans(duration);",
	}

	for _, index := range indexes {
		if err := d.DB.Exec(index).Error; err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	log.Println("Database indexes created successfully")
	return nil
}


func (d *Database) SeedData() error {
	log.Println("Seeding database with sample data...")

	
	var count int64
	d.DB.Model(&models.Product{}).Count(&count)
	if count > 0 {
		log.Println("Database already contains data, skipping seed")
		return nil
	}

	
products := []models.Product{
    {
        Name:        "Premium Software License",
        Description: "Annual license for premium software package",
        Price:       99.99,
        Type:        models.ProductTypeDigital,
        DigitalDetails: &models.DigitalProductDetails{
            
            FileSize:     104857600,
            DownloadLink: "https://url.com",
        },
    },
    {
        Name:        "Deluxe Physical Widget",
        Description: "High-end physical widget with premium materials",
        Price:       149.99,
        Type:        models.ProductTypePhysical,
        PhysicalDetails: &models.PhysicalProductDetails{
            Weight:     2.5,
            Dimensions: "10x20x5 cm",
        },
    },
    {
        Name:        "Gold Subscription",
        Description: "Monthly gold-level subscription",
        Price:       0, 
        Type:        models.ProductTypeSubscription,
        SubscriptionDetails: &models.SubscriptionProductDetails{
            SubscriptionPeriod: "monthly",
            RenewalPrice:       19.99,
        },
    },
}

	for _, product := range products {
		if err := d.DB.Create(&product).Error; err != nil {
			return fmt.Errorf("failed to seed product: %w", err)
		}

		
		if product.Type == models.ProductTypeSubscription {
			plans := []models.SubscriptionPlan{
				{
					ProductID: product.ID,
					PlanName:  "Basic Plan",
					Duration:  30,
					Price:     9.99,
				},
				{
					ProductID: product.ID,
					PlanName:  "Premium Plan",
					Duration:  365,
					Price:     99.99,
				},
			}

			for _, plan := range plans {
				if err := d.DB.Create(&plan).Error; err != nil {
					return fmt.Errorf("failed to seed subscription plan: %w", err)
				}
			}
		}
	}

	log.Println("Database seeded successfully")
	return nil
}


func (d *Database) HealthCheck() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}


func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
