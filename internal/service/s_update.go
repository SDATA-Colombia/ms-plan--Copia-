package service

import (
	"context"
	"time"

	"github.com/SDATA-Colombia/ms-plan/internal/model/entity"
	"github.com/SDATA-Colombia/ms-plan/pkg/logger"
	"github.com/SDATA-Colombia/ms-plan/pkg/pulsar"

	"go.mongodb.org/mongo-driver/v2/bson"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// BaseService está en el mismo package service.
// repository ya está importado en base.go.
// T es el genérico declarado en BaseService[T].
// repository.UpdateByID() y UpdateMany() ya son accesibles a través de s.repository.
// En Go cada archivo importa únicamente lo que utiliza, aunque pertenezcan al mismo package.

// ----------------------------------------------------------------------
// UPDATE
// ----------------------------------------------------------------------

// UpdateByID actualiza una entidad por su identificador.
func (s *BaseService) UpdateByID(
	ctx context.Context,
	id bson.ObjectID,
	entity *entity.Plan,

) (*entity.Plan, error) {

	// ------------------------------------------------------------------
	// VALIDACIONES
	// ------------------------------------------------------------------

	if entity == nil {
		return nil, status.Error(
			codes.InvalidArgument,
			"Plan data is required",
		)
	}

	if entity.PlanTitle == "" {
		return nil, status.Error(
			codes.InvalidArgument,
			"Plan title is required",
		)
	}

	if entity.PlanSubtitle == "" {
		return nil, status.Error(
			codes.InvalidArgument,
			"Plan subtitle is required",
		)
	}

	// Validaciones locales.
	update := bson.M{
		"plan_id_company":           entity.PlanIDCompany,
		"plan_accent_color":         entity.PlanAccentColor,
		"plan_accent_text_color":    entity.PlanAccentTextColor,
		"plan_title":                entity.PlanTitle,
		"plan_subtitle":             entity.PlanSubtitle,
		"plan_description_tinymce":  entity.PlanDescriptionTinyMCE,
		"plan_type":                 entity.PlanType,
		"plan_currency_price":       entity.PlanCurrencyPrice,
		"plan_menu":                 entity.PlanMenu,
		"plan_description_items":    entity.PlanDescriptionItems,
		"plan_days":                 entity.PlanDays,
		"plan_offices":              entity.PlanOffices,
		"plan_users":                entity.PlanUsers,
		"plan_total_sales":          entity.PlanTotalSales,
		"plan_price":                entity.PlanPrice,
		"plan_featured":             entity.PlanFeatured,
		"plan_custom":               entity.PlanCustom,
		"active_plan":               entity.ActivePlan,
		"updatedAt":                 time.Now(), // Recomendado para refrescar automáticamente la fecha de modificación
	}

	s.applyUpdateDefaults(update)

	//----------------------------------------------------------------------
	// REGLA DE NEGOCIO
	//----------------------------------------------------------------------

	// Si el nuevo plan será destacado,
	// todos los demás planes destacados dejan de serlo.
	if entity.PlanFeatured {

		if err := s.repository.UpdateMany(
			ctx,
			bson.M{
				"plan_featured": true,
				"_id": bson.M{
					"$ne": id,
				},
			},
			bson.M{
				"plan_featured": false,
			},
	
		); err != nil {

			logger.Error(err.Error())

			return nil, status.Error(
				codes.Internal,
				"internal server error",
			)
		}
	}

	//----------------------------------------------------------------------
	// PERSISTENCIA
	//----------------------------------------------------------------------

	if err := s.repository.UpdateByID(ctx, id, update); err != nil {
		return nil, err
	}

	//----------------------------------------------------------------------
	// INVALIDAR CACHÉ
	//----------------------------------------------------------------------

	if err := s.bumpCacheVersion(
		ctx,
		"plan",
		"global",
	
	); err != nil {
		logger.Error(err.Error())
	}

	//----------------------------------------------------------------------
	// PUBLICAR EVENTO
	//----------------------------------------------------------------------

	event := pulsar.NewEvent(
		pulsar.TopicPlanUpdated,
		"ms-plan",
		entity,
	)

	event.WithCorrelation(id.Hex())

	if err := pulsar.Publish(
		ctx,
		pulsar.TopicPlanUpdated,
		event,
	
	); err != nil {

        logger.Error(err.Error())
	}

	return s.repository.FindByID(
		ctx,
		id,
	)
}