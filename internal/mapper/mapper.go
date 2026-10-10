package mapper

// Mapper: Puente entre entity (Mongo) y proto (red).
// NO contiene lógica de negocio, NO decide rutas, NO hace I/O.

import (
	planpb "github.com/SDATA-Colombia/shared-go-proto/plan/v1"

	"github.com/SDATA-Colombia/ms-plan/internal/model/entity"

	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ProtoToEntity convierte un PlanEntity (gRPC)
// en una entidad de MongoDB.

// ----------------------------------------------------------------------
// Proto -> Entity
// ----------------------------------------------------------------------

func ProtoToEntity(proto *planpb.PlanEntity) (*entity.Plan, error) {

	var objectID bson.ObjectID

	if proto.GetId() != "" {
		id, err := bson.ObjectIDFromHex(proto.GetId())
		if err != nil {
			return nil, err
		}
		objectID = id
	}

	var idCompany bson.ObjectID

	if proto.GetPlanIdCompany() != "" {
		id, err := bson.ObjectIDFromHex(proto.GetPlanIdCompany())
		if err != nil {
			return nil, err
		}
		idCompany = id
	}

	plan := &entity.Plan{
		ID:                     objectID,
		PlanIDCompany:          idCompany,
		PlanAccentColor:        proto.GetPlanAccentColor(),
		PlanAccentTextColor:    proto.GetPlanAccentTextColor(),
		PlanTitle:              proto.GetPlanTitle(),
		PlanSubtitle:           proto.GetPlanSubtitle(),
		PlanDescriptionTinyMCE: proto.GetPlanDescriptionTinymce(),
		PlanType:               proto.GetPlanType(),
		PlanCurrencyPrice:      proto.GetPlanCurrencyPrice(),
		PlanMenu:               proto.GetPlanMenu(),
		PlanDescriptionItems:   proto.GetPlanDescriptionItems(),

		PlanDays:       proto.GetPlanDays(),
		PlanOffices:    proto.GetPlanOffices(),
		PlanUsers:      proto.GetPlanUsers(),
		PlanTotalSales: proto.GetPlanTotalSales(),
		PlanPrice:      proto.GetPlanPrice(),

		PlanFeatured: proto.GetPlanFeatured(),
		PlanCustom:   proto.GetPlanCustom(),
		ActivePlan:   proto.GetActivePlan(),
	}

	if proto.GetCreatedAt() != nil {
		plan.CreatedAt = proto.GetCreatedAt().AsTime()
	}

	if proto.GetUpdatedAt() != nil {
		plan.UpdatedAt = proto.GetUpdatedAt().AsTime()
	}

	if proto.GetDeletedAt() != nil {
		t := proto.GetDeletedAt().AsTime()
		plan.DeletedAt = &t
	}

	return plan, nil
}

// EntityToProto convierte una entidad MongoDB
// en PlanEntity (gRPC).

// ----------------------------------------------------------------------
// Entity -> Proto
// ----------------------------------------------------------------------

func EntityToProto(plan *entity.Plan) *planpb.PlanEntity {
	if plan == nil {
		return nil
	}

	var id string
	if !plan.ID.IsZero() {
		id = plan.ID.Hex()
	}

	var planIDCompany string
	if !plan.PlanIDCompany.IsZero() {
		planIDCompany = plan.PlanIDCompany.Hex()
	}

	var deletedAt *timestamppb.Timestamp
	if plan.DeletedAt != nil {
		deletedAt = timestamppb.New(*plan.DeletedAt)
	}

	return &planpb.PlanEntity{
		Id:                     id,
		PlanIdCompany:          planIDCompany,
		PlanAccentColor:        plan.PlanAccentColor,
		PlanAccentTextColor:    plan.PlanAccentTextColor,
		PlanTitle:              plan.PlanTitle,
		PlanSubtitle:           plan.PlanSubtitle,
		PlanDescriptionTinymce: plan.PlanDescriptionTinyMCE,
		PlanType:               plan.PlanType,
		PlanCurrencyPrice:      plan.PlanCurrencyPrice,
		PlanMenu:               plan.PlanMenu,
		PlanDescriptionItems:   plan.PlanDescriptionItems,

		PlanDays:       plan.PlanDays,
		PlanOffices:    plan.PlanOffices,
		PlanUsers:      plan.PlanUsers,
		PlanTotalSales: plan.PlanTotalSales,
		PlanPrice:      plan.PlanPrice,

		PlanFeatured: plan.PlanFeatured,
		PlanCustom:   plan.PlanCustom,
		ActivePlan:   plan.ActivePlan,

		CreatedAt: timestamppb.New(plan.CreatedAt),
		UpdatedAt: timestamppb.New(plan.UpdatedAt),
		DeletedAt: deletedAt,
	}
}

// EntitiesToProto convierte una lista de entidades
// en una lista de PlanEntity.
func EntitiesToProto(bases []*entity.Plan) []*planpb.PlanEntity {

	result := make([]*planpb.PlanEntity, 0, len(bases))

	for _, plan := range bases {
		result = append(result, EntityToProto(plan))
	}

	return result
}
