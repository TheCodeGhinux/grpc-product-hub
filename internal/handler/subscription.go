// Package handler contains gRPC server implementations for handling subscription plan requests.
package handler

import (
	"context"
	"grpc-product/internal/types"
	pb "grpc-product/pkg/pb/api/proto"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)


type SubscriptionPlanHandler struct {
	pb.UnimplementedSubscriptionPlanServiceServer
	subscriptionService types.SubscriptionPlanService
}


func NewSubscriptionPlanHandler(subscriptionService types.SubscriptionPlanService) *SubscriptionPlanHandler {
	return &SubscriptionPlanHandler{
		subscriptionService: subscriptionService,
	}
}

func (h *SubscriptionPlanHandler) CreateSubscriptionPlan(ctx context.Context, req *pb.CreateSubscriptionPlanRequest) (*pb.SubscriptionPlanResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	if req.ProductId == "" {
		return nil, status.Error(codes.InvalidArgument, "product ID is required")
	}

	
	productID, err := uuid.Parse(req.ProductId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid product ID format")
	}

	
	serviceReq := &types.CreateSubscriptionPlanRequest{
		ProductID: productID,
		PlanName:  req.PlanName,
		Duration:  req.Duration,
		Price:     req.Price,
	}

	plan, err := h.subscriptionService.CreateSubscriptionPlan(ctx, serviceReq)
	if err != nil {
		return nil, convertServiceErrorToGRPCError(err)
	}

	
	protoPlan := convertModelSubscriptionPlanToProto(plan)
	return &pb.SubscriptionPlanResponse{Plan: protoPlan}, nil
}


func (h *SubscriptionPlanHandler) GetSubscriptionPlan(ctx context.Context, req *pb.GetSubscriptionPlanRequest) (*pb.SubscriptionPlanResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "subscription plan ID is required")
	}

	
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid subscription plan ID format")
	}

	
	plan, err := h.subscriptionService.GetSubscriptionPlan(ctx, id)
	if err != nil {
		return nil, convertServiceErrorToGRPCError(err)
	}

	
	protoPlan := convertModelSubscriptionPlanToProto(plan)
	return &pb.SubscriptionPlanResponse{Plan: protoPlan}, nil
}


func (h *SubscriptionPlanHandler) GetProductSubscriptionPlans(ctx context.Context, req *pb.GetProductSubscriptionPlansRequest) (*pb.ListSubscriptionPlansResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	if req.ProductId == "" {
		return nil, status.Error(codes.InvalidArgument, "product ID is required")
	}

	
	productID, err := uuid.Parse(req.ProductId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid product ID format")
	}

	
	plans, err := h.subscriptionService.GetProductSubscriptionPlans(ctx, productID)
	if err != nil {
		return nil, convertServiceErrorToGRPCError(err)
	}

	
	var protoPlans []*pb.SubscriptionPlan
	for _, plan := range plans {
		protoPlans = append(protoPlans, convertModelSubscriptionPlanToProto(plan))
	}

	return &pb.ListSubscriptionPlansResponse{Plans: protoPlans}, nil
}


func (h *SubscriptionPlanHandler) UpdateSubscriptionPlan(ctx context.Context, req *pb.UpdateSubscriptionPlanRequest) (*pb.SubscriptionPlanResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "subscription plan ID is required")
	}

	if req.PlanName == "" {
		return nil, status.Error(codes.InvalidArgument, "plan name is required")
	}

	
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid subscription plan ID format")
	}

	
	serviceReq := &types.UpdateSubscriptionPlanRequest{
		PlanName: req.PlanName,
		Duration: req.Duration,
		Price:    req.Price,
	}

	
	plan, err := h.subscriptionService.UpdateSubscriptionPlan(ctx, id, serviceReq)
	if err != nil {
		return nil, convertServiceErrorToGRPCError(err)
	}

	
	protoPlan := convertModelSubscriptionPlanToProto(plan)
	return &pb.SubscriptionPlanResponse{Plan: protoPlan}, nil
}


func (h *SubscriptionPlanHandler) DeleteSubscriptionPlan(ctx context.Context, req *pb.DeleteSubscriptionPlanRequest) (*emptypb.Empty, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "subscription plan ID is required")
	}

	
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid subscription plan ID format")
	}

	
	if err := h.subscriptionService.DeleteSubscriptionPlan(ctx, id); err != nil {
		return nil, convertServiceErrorToGRPCError(err)
	}

	return &emptypb.Empty{}, nil
}
