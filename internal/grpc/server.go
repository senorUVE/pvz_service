package grpc

import (
	"context"
	"net"
	"time"

	pbv1 "github.com/senorUVE/pvz_service/internal/generated"
	"github.com/senorUVE/pvz_service/internal/repository"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const maxLimit = 30

type Server struct {
	pbv1.UnimplementedPVZServiceServer
	repo *repository.Repository
}

func NewServer(repo *repository.Repository) *Server {
	return &Server{repo: repo}
}

func (s *Server) GetPVZList(ctx context.Context, _ *pbv1.GetPVZListRequest) (*pbv1.GetPVZListResponse, error) {
	pvzList, err := s.repo.GetPvz(ctx, time.Time{}, time.Time{}, 1, maxLimit)
	if err != nil {
		return nil, err
	}

	resp := &pbv1.GetPVZListResponse{
		Pvzs: make([]*pbv1.PVZ, 0, len(pvzList)),
	}
	for _, p := range pvzList {
		resp.Pvzs = append(resp.Pvzs, &pbv1.PVZ{
			Id:               p.PVZ.Id.String(),
			RegistrationDate: timestamppb.New(p.PVZ.RegistrationDate),
			City:             p.PVZ.City,
		})
	}
	return resp, nil
}

func StartGrpcServer(repo *repository.Repository) error {
	lis, err := net.Listen("tcp", ":3000")
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	pbv1.RegisterPVZServiceServer(grpcServer, NewServer(repo))
	reflection.Register(grpcServer)
	return grpcServer.Serve(lis)
}
