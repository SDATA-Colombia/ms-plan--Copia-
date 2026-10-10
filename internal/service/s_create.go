package service

import (
	"context"
	"errors"

	"github.com/SDATA-Colombia/ms-plan/internal/model/entity"
	"github.com/SDATA-Colombia/ms-plan/pkg/logger"
	"github.com/SDATA-Colombia/ms-plan/pkg/pulsar"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"golang.org/x/sync/errgroup"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// BaseService está en el mismo package service.
// repository ya está importado en base.go.
// T es el genérico declarado en BaseService[T].
// repository.Create() e InsertMany() ya son accesibles a través de s.repository.
// En Go cada archivo importa únicamente lo que utiliza, aunque pertenezcan al mismo package.

// ----------------------------------------------------------------------
// CREATE
// ----------------------------------------------------------------------

// Create registra una nueva entidad.
func (s *BaseService) Create(
	ctx context.Context,
	plan *entity.Plan,

) error {

	s.applyCreateDefaults(plan)

	//----------------------------------------------------------------------
	// EJECUCIÓN PARALELA
	//----------------------------------------------------------------------

	g, ctxGroup := errgroup.WithContext(ctx)

	// ------------------------------------------------------------------
	// VALIDACIONES
	// ------------------------------------------------------------------

	// Validaciones locales.
	g.Go(func() error {

		if plan.PlanTitle == "" {
			return status.Error(
				codes.InvalidArgument,
				"Plan title is required",
			)
		}

		if plan.PlanSubtitle == "" {
			return status.Error(
				codes.InvalidArgument,
				"Plan subtitle is required",
			)
		}

		return nil
	})

	// Verificar que el plan no exista.
	g.Go(func() error {

		existing, err := s.FindByField(
			ctxGroup,
			"plan_title",
			plan.PlanTitle,
		)

		if err != nil {

			// Que no exista el plan NO es un error.
			if errors.Is(err, mongo.ErrNoDocuments) {
				return nil
			}

			return err
		}

		if existing != nil {
			return status.Error(
				codes.AlreadyExists,
				"The plan is already registered",
			)
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}

	//----------------------------------------------------------------------
	// REGLA DE NEGOCIO
	//----------------------------------------------------------------------

	// Si el nuevo plan será destacado,
	// todos los demás planes destacados dejan de serlo.
	if plan.PlanFeatured {

		if err := s.repository.UpdateMany(
			ctx,
			bson.M{
				"plan_featured": true,
			},
			bson.M{
				"plan_featured": false,
			},

		); err != nil {

			logger.Error(err.Error())

			return status.Error(
				codes.Internal,
				"internal server error",
			)
		}
	}

	//----------------------------------------------------------------------
	// PERSISTENCIA
	//----------------------------------------------------------------------

	if err := s.repository.Create(ctx, plan); err != nil {

		if mongo.IsDuplicateKeyError(err) {
			return status.Error(
				codes.AlreadyExists,
				"The plan is already registered",
			)
		}

		logger.Error(err.Error())

		return status.Error(
			codes.Internal,
			"internal server error",
		)
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
		pulsar.TopicPlanCreated,
		"ms-plan",
		plan,
	)

	event.WithCorrelation(plan.ID.Hex())

	if err := pulsar.Publish(
		ctx,
		pulsar.TopicPlanCreated,
		event,
	
	); err != nil {

		logger.Error(err.Error())

		// return status.Error(
		// 	codes.Internal,
		// 	"internal server error",
		// )
	} else {

		logger.Info(
			"Evento publicado",
			"topic", pulsar.TopicPlanCreated,
			"entity_id", plan.ID.Hex(),
		)
	}

	return nil
}
