package controller

import (
	"context"
	"strings"

	planpb "github.com/SDATA-Colombia/shared-go-proto/plan/v1"

	"github.com/SDATA-Colombia/ms-plan/internal/mapper"

	"go.mongodb.org/mongo-driver/v2/bson"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (c *PlanController) GetPlan(
	ctx context.Context,
	req *planpb.GetPlanReq,

) (*planpb.GetPlanRes, error) {

	idStr := strings.TrimSpace(req.GetId())
    if idStr == "" {
        return nil, status.Error(codes.InvalidArgument, "ID is required")
    }

    id, err := bson.ObjectIDFromHex(idStr)
    if err != nil {
        return nil, status.Error(codes.InvalidArgument, "Invalid ID format")
    }

	plan, err := c.service.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if plan == nil {
        return nil, status.Error(codes.NotFound, "Plan not found")
    }

	return &planpb.GetPlanRes{
		Success: true,
		Message: "Plan successfully queried by ID",
		Data:    mapper.EntityToProto(plan),
	}, nil
}

func (c *PlanController) ListPlan(
	ctx context.Context,
	req *planpb.ListPlanReq,

) (*planpb.ListPlanRes, error) {

	entities, err := c.service.FindManyPaginated(
		ctx,
		bson.M{},
		int64(req.GetPage()),
		int64(req.GetLimit()),
	)
	if err != nil {
		return nil, err
	}

	total, err := c.service.Count(
		ctx,
		bson.M{},
	)
	if err != nil {
		return nil, err
	}

	return &planpb.ListPlanRes{
		Success: true,
		Message: "Plan consultadas correctamente",
		Data:    mapper.EntitiesToProto(entities),
		Total:   total,
	}, nil
}
