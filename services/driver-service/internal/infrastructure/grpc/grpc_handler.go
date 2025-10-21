package grpc

import (
	"context"
	"ride-sharing/services/driver-service/internal/service"

	pb "ride-sharing/shared/proto/driver"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcHandler struct {
	pb.UnimplementedDriverServiceServer
	service *service.Service
}

func NewGrpcHandler(server *grpc.Server, service *service.Service) *grpcHandler {
	handler := &grpcHandler{
		service: service,
	}
	pb.RegisterDriverServiceServer(server, handler)

	return handler
}

func (h *grpcHandler) RegisterDriver(context.Context, *pb.RegisterDriverRequest) (*pb.RegisterDriverResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterDriver not implemented")
}
func (h *grpcHandler) UnregisterDriver(context.Context, *pb.RegisterDriverRequest) (*pb.RegisterDriverResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UnregisterDriver not implemented")
}
