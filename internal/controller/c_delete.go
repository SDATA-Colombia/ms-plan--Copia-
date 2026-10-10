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

func (c *PlanController) DeletePlan(
	ctx context.Context,
	req *planpb.DeletePlanReq,

) (*planpb.DeletePlanRes, error) {
	
	idStr := strings.TrimSpace(req.GetId())
    if idStr == "" {
        return nil, status.Error(codes.InvalidArgument, "ID is required")
    }

    id, err := bson.ObjectIDFromHex(idStr)
    if err != nil {
        return nil, status.Error(codes.InvalidArgument, "Invalid ID format")
    }

	if err := c.service.DeleteByID(ctx, id); err != nil {

		if errors.Is(err, mongo.ErrNoDocuments) {
            return nil, status.Error(codes.NotFound, "Plan not found")
        }

		return nil, err
	}

	return &planpb.DeletePlanRes{
		Success: true,
		Message: "A plan has been successfully deleted",
	}, nil
}

func (c *PlanController) DeleteFeaturePlan(
	ctx context.Context,
	req *planpb.DeleteFeaturePlanReq,

) (*planpb.DeleteFeaturePlanRes, error) {
	
	id := req.GetId()
	// if id <= 0 {
	// 	return nil, status.Error(codes.InvalidArgument, "A feature ID is required")
	// }

	idStrP := strings.TrimSpace(req.GetPlanId())
    if idStrP == "" {
        return nil, status.Error(codes.InvalidArgument, "Plan ID is required")
    }

    idP, err := bson.ObjectIDFromHex(idStrP)
    if err != nil {
        return nil, status.Error(codes.InvalidArgument, "Invalid ID format")
    }

	updatedPlan, err := c.service.DeleteByFeatureID(
		ctx,
		id,
		idP,
	)
	if err != nil {

		if errors.Is(err, mongo.ErrNoDocuments) {
            return nil, status.Error(codes.NotFound, "Plan not found")
        }

		return nil, err
	}

	return &planpb.DeleteFeaturePlanRes{
		Success: true,
		Message: "A feature has been removed",
		Data:    mapper.EntityToProto(updatedPlan),
	}, nil
}
