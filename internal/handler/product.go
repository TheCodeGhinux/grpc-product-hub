package handler

import (
	"context"
	"errors"
	"fmt"
	"grpc-product/internal/models"
	"grpc-product/internal/types"
	pb "grpc-product/pkg/pb/api/proto"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)


type ProductHandler struct {
	pb.UnimplementedProductServiceServer
	productService types.ProductService
}


func NewProductHandler(svc types.ProductService) *ProductHandler {
	return &ProductHandler{productService: svc}
}


func (h *ProductHandler) convertModelToProtoType(t models.ProductType) pb.ProductType {
	switch t {
	case models.ProductTypeDigital:
		return pb.ProductType_digital
	case models.ProductTypePhysical:
		return pb.ProductType_physical
	case models.ProductTypeSubscription:
		return pb.ProductType_subscription
	default:
		return pb.ProductType_PRODUCT_TYPE_UNSPECIFIED
	}
}


func (h *ProductHandler) convertProtoToModelType(pt pb.ProductType) models.ProductType {
	switch pt {
	case pb.ProductType_digital:
		return models.ProductTypeDigital
	case pb.ProductType_physical:
		return models.ProductTypePhysical
	case pb.ProductType_subscription:
		return models.ProductTypeSubscription
	default:
		return models.ProductTypeDigital
	}
}


func (h *ProductHandler) convertProductToProto(p *models.Product) (*pb.Product, error) {
	if p == nil {
		return nil, fmt.Errorf("product is nil")
	}

	msg := &pb.Product{
		Id:          p.ID.String(),
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Type:        h.convertModelToProtoType(p.Type),
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}

	switch p.Type {
	case models.ProductTypeDigital:
		if d := p.DigitalDetails; d != nil {
			msg.ProductDetails = &pb.Product_DigitalDetails{DigitalDetails: &pb.DigitalProductDetails{
				FileSize:     d.FileSize,
				DownloadLink: d.DownloadLink,
			}}
		}
	case models.ProductTypePhysical:
		if d := p.PhysicalDetails; d != nil {
			msg.ProductDetails = &pb.Product_PhysicalDetails{PhysicalDetails: &pb.PhysicalProductDetails{
				Weight:     d.Weight,
				Dimensions: d.Dimensions,
			}}
		}
	case models.ProductTypeSubscription:
		if d := p.SubscriptionDetails; d != nil {
			msg.ProductDetails = &pb.Product_SubscriptionDetails{SubscriptionDetails: &pb.SubscriptionProductDetails{
				SubscriptionPeriod: d.SubscriptionPeriod,
				RenewalPrice:       d.RenewalPrice,
			}}
		}
	}

	return msg, nil
}


func (h *ProductHandler) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.ProductResponse, error) {
	if req == nil || req.Name == "" || req.Price < 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid create request")
	}

	svcReq := &types.CreateProductRequest{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Type:        h.convertProtoToModelType(req.Type),
	}
	switch req.Type {
	case pb.ProductType_digital:
		svcReq.DigitalDetails = &models.DigitalProductDetails{
			FileSize:     req.GetDigitalDetails().FileSize,
			DownloadLink: req.GetDigitalDetails().DownloadLink,
		}
	case pb.ProductType_physical:
		svcReq.PhysicalDetails = &models.PhysicalProductDetails{
			Weight:     req.GetPhysicalDetails().Weight,
			Dimensions: req.GetPhysicalDetails().Dimensions,
		}
	case pb.ProductType_subscription:
		svcReq.SubscriptionDetails = &models.SubscriptionProductDetails{
			SubscriptionPeriod: req.GetSubscriptionDetails().SubscriptionPeriod,
			RenewalPrice:       req.GetSubscriptionDetails().RenewalPrice,
		}
	}

	prod, err := h.productService.CreateProduct(ctx, svcReq)
	if err != nil {
		return nil, convertServiceErrorToGRPCError(err)
	}

	out, err := h.convertProductToProto(prod)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err)
	}
	return &pb.ProductResponse{Product: out}, nil
}


func (h *ProductHandler) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.ProductResponse, error) {
	if req == nil || req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "product ID is required")
	}
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid product ID format")
	}

	prod, err := h.productService.GetProduct(ctx, id)
	if err != nil {
		return nil, convertServiceErrorToGRPCError(err)
	}

	out, err := h.convertProductToProto(prod)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err)
	}
	return &pb.ProductResponse{Product: out}, nil
}


func (h *ProductHandler) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.ProductResponse, error) {
	if req == nil || req.Id == "" || req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "product ID and name are required")
	}
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid product ID format")
	}

	svcReq := &types.UpdateProductRequest{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
	}
	if d := req.GetDigitalDetails(); d != nil {
		svcReq.DigitalDetails = &models.DigitalProductDetails{FileSize: d.FileSize, DownloadLink: d.DownloadLink}
	}
	if d := req.GetPhysicalDetails(); d != nil {
		svcReq.PhysicalDetails = &models.PhysicalProductDetails{Weight: d.Weight, Dimensions: d.Dimensions}
	}
	if d := req.GetSubscriptionDetails(); d != nil {
		svcReq.SubscriptionDetails = &models.SubscriptionProductDetails{SubscriptionPeriod: d.SubscriptionPeriod, RenewalPrice: d.RenewalPrice}
	}

	prod, err := h.productService.UpdateProduct(ctx, id, svcReq)
	if err != nil {
		return nil, convertServiceErrorToGRPCError(err)
	}

	out, err := h.convertProductToProto(prod)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err)
	}
	return &pb.ProductResponse{Product: out}, nil
}


func (h *ProductHandler) DeleteProduct(ctx context.Context, req *pb.DeleteProductRequest) (*emptypb.Empty, error) {
	if req == nil || req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "product ID is required")
	}
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid product ID format")
	}
	if err := h.productService.DeleteProduct(ctx, id); err != nil {
		return nil, convertServiceErrorToGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}


func (h *ProductHandler) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	serviceReq := &types.ListProductsRequest{Page: int(req.Page), PageSize: int(req.PageSize)}
	if req.TypeFilter != pb.ProductType_PRODUCT_TYPE_UNSPECIFIED {
		t := h.convertProtoToModelType(req.TypeFilter)
		serviceReq.Type = &t
	}
	if serviceReq.Page < 1 {
		serviceReq.Page = 1
	}
	if serviceReq.PageSize < 1 {
		serviceReq.PageSize = 10
	}

	resp, err := h.productService.ListProducts(ctx, serviceReq)
	if err != nil {
		return nil, convertServiceErrorToGRPCError(err)
	}

	var items []*pb.Product
	for _, p := range resp.Products {
		msg, err := h.convertProductToProto(p)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "conversion error: %v", err)
		}
		items = append(items, msg)
	}
	return &pb.ListProductsResponse{Products: items, TotalCount: int32(resp.TotalCount), Page: int32(resp.Page), PageSize: int32(resp.PageSize)}, nil
}

func convertModelSubscriptionPlanToProto(plan *models.SubscriptionPlan) *pb.SubscriptionPlan {
	if plan == nil {
		return nil
	}

	return &pb.SubscriptionPlan{
		Id:        plan.ID.String(),
		ProductId: plan.ProductID.String(),
		PlanName:  plan.PlanName,
		Duration:  plan.Duration,
		Price:     plan.Price,
		CreatedAt: timestamppb.New(plan.CreatedAt),
		UpdatedAt: timestamppb.New(plan.UpdatedAt),
	}
}

func convertServiceErrorToGRPCError(err error) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	
	switch {
	case errors.Is(err, fmt.Errorf("not found")) ||
		containsAny(errMsg, "not found", "does not exist"):
		return status.Error(codes.NotFound, errMsg)

	case containsAny(errMsg, "validation failed", "invalid", "required", "cannot be"):
		return status.Error(codes.InvalidArgument, errMsg)

	case containsAny(errMsg, "already exists", "duplicate", "conflict"):
		return status.Error(codes.AlreadyExists, errMsg)

	case containsAny(errMsg, "permission denied", "unauthorized", "access denied"):
		return status.Error(codes.PermissionDenied, errMsg)

	case containsAny(errMsg, "timeout", "deadline", "context deadline exceeded"):
		return status.Error(codes.DeadlineExceeded, errMsg)

	case containsAny(errMsg, "unavailable", "connection", "network"):
		return status.Error(codes.Unavailable, errMsg)

	default:
		
		return status.Error(codes.Internal, fmt.Sprintf("internal server error: %s", errMsg))
	}
}


func containsAny(text string, substrings ...string) bool {
	for _, substr := range substrings {
		if len(text) >= len(substr) {
			for i := 0; i <= len(text)-len(substr); i++ {
				if text[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}
