// Package server handles server
package server

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	_ "github.com/joho/godotenv/autoload"
	"google.golang.org/grpc"

	"grpc-product/internal/database"
	"grpc-product/internal/handler"
	"grpc-product/internal/repository"
	service "grpc-product/internal/services"
	pb "grpc-product/pkg/pb/api/proto"
)

func StartGRPCServer() {
	// Load port from env
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil || port == 0 {
		log.Fatalf("Invalid or missing PORT environment variable: %v", err)
	}

	// Initialize DB
	dbConfig := database.GetDatabaseConfig()
	db, err := database.NewDatabase(dbConfig)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	// Auto migrate schema
	if err := db.AutoMigrate(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// Seed data (optional)
	if err := db.SeedData(); err != nil {
		log.Fatalf("Seeding failed: %v", err)
	}

	// Initialize layers: Repository -> Service -> Handler
	productRepo := repository.NewProductRepository(db.DB)
	subscriptionRepo := repository.NewSubscriptionPlanRepository(db.DB)

	productService := service.NewProductService(productRepo)
	subscriptionService := service.NewSubscriptionPlanService(subscriptionRepo)

	productHandler := handler.NewProductHandler(productService)
	subscriptionHandler := handler.NewSubscriptionPlanHandler(subscriptionService)

	// Setup gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	// Register gRPC services
	pb.RegisterProductServiceServer(grpcServer, productHandler)
	pb.RegisterSubscriptionPlanServiceServer(grpcServer, subscriptionHandler)

	log.Printf("✅ gRPC server listening on port %d", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}
}
