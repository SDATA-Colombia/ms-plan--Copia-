package controller

import (
	"context"
	"errors"
	"strings"

	planpb "github.com/SDATA-Colombia/shared-go-proto/plan/v1"

	"github.com/SDATA-Colombia/ms-plan/internal/mapper"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (c *PlanController) UpdatePlan(
	ctx context.Context,
	req *planpb.UpdatePlanReq,

) (*planpb.UpdatePlanRes, error) {

	idStr := strings.TrimSpace(req.GetId())
    if idStr == "" {
        return nil, status.Error(codes.InvalidArgument, "ID is required")
    }

    id, err := bson.ObjectIDFromHex(idStr)
    if err != nil {
        return nil, status.Error(codes.InvalidArgument, "Invalid ID format")
    }

	plan, err := mapper.ProtoToEntity(req.GetData())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	updatedPlan, err := c.service.UpdateByID(
		ctx,
		id,
		plan,
	)
	if err != nil {

		if errors.Is(err, mongo.ErrNoDocuments) {
            return nil, status.Error(codes.NotFound, "Plan not found")
        }

		return nil, err
	}

	return &planpb.UpdatePlanRes{
		Success: true,
		Message: "The plan has been successfully updated",
		Data:    mapper.EntityToProto(updatedPlan),
	}, nil
}
