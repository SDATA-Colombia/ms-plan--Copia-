package controller

import (
	"context"

	planpb "github.com/SDATA-Colombia/shared-go-proto/plan/v1"

	"github.com/SDATA-Colombia/ms-plan/internal/mapper"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (c *PlanController) CreatePlan(
	ctx context.Context,
	req *planpb.CreatePlanReq,

) (*planpb.CreatePlanRes, error) {

	plan, err := mapper.ProtoToEntity(req.GetData())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := c.service.Create(ctx, plan); err != nil {
		return nil, err
	}

	return &planpb.CreatePlanRes{
		Success: true,
		Message: "A plan has been successfully registered",
		Data:    mapper.EntityToProto(plan),
	}, nil
}
