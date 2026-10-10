package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/SDATA-Colombia/ms-plan/internal/model/entity"
	"github.com/SDATA-Colombia/ms-plan/pkg/logger"
	"github.com/SDATA-Colombia/ms-plan/pkg/pulsar"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// BaseService está en el mismo package service.
// repository ya está importado en base.go.
// T es el genérico declarado en BaseService[T].
// repository.DeleteByID() y DeleteMany() ya son accesibles a través de s.repository.
// En Go cada archivo importa únicamente lo que utiliza, aunque pertenezcan al mismo package.

// ----------------------------------------------------------------------
// DELETE
// ----------------------------------------------------------------------

// DeleteByID elimina físicamente una entidad.
func (s *BaseService) DeleteByID(
	ctx context.Context,
	id bson.ObjectID,

) error {

	// Aquí irán las validaciones.

	// ------------------------------------------------------------------
	// OBTENER ENTIDAD ANTES DE ELIMINAR (Para compensación y eventos)
	// ------------------------------------------------------------------

	plan, err := s.repository.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return status.Error(
				codes.NotFound,
				"Plan not found",
			)
		}

		logger.Error(err.Error())

		return status.Error(
			codes.Internal,
			"Internal server error",
		)
	}

	//----------------------------------------------------------------------
	// PERSISTENCIA
	//----------------------------------------------------------------------

	if err := s.repository.DeleteByID(ctx, id); err != nil {

		if errors.Is(err, mongo.ErrNoDocuments) {
			return status.Error(
				codes.NotFound,
				"Plan not found",
			)
		}

		logger.Error(err.Error())

		return status.Error(
			codes.Internal,
			"Internal server error",
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
		pulsar.TopicPlanDeleted,
		"ms-plan",
		plan,
	)

	event.WithCorrelation(plan.ID.Hex())

	if err := pulsar.Publish(
		ctx,
		pulsar.TopicPlanDeleted,
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
			"topic", pulsar.TopicPlanDeleted,
			"entity_id", plan.ID.Hex(),
		)
	}

	return nil
}

// DeleteByFeatureID elimina una característica de una entidad.
func (s *BaseService) DeleteByFeatureID(
	ctx context.Context,
	id int32,
	idP bson.ObjectID,

) (*entity.Plan, error) {

	// Aquí irán las validaciones.

	//----------------------------------------------------------------------
	// OBTENER PLAN
	//----------------------------------------------------------------------

	plan, err := s.repository.FindByID(ctx, idP)
	if err != nil {

		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, status.Error(
				codes.NotFound,
				"Plan not found",
			)
		}

		logger.Error(err.Error())

		return nil, status.Error(
			codes.Internal,
			"Internal server error",
		)
	}

	if plan == nil {
		return nil, status.Error(
			codes.NotFound,
			"Plan not found",
		)
	}

	//----------------------------------------------------------------------
	// CONVERTIR JSON STRING → ARRAY
	//----------------------------------------------------------------------

	var features []map[string]json.RawMessage

	if plan.PlanDescriptionItems != "" {

		if err := json.Unmarshal(
			[]byte(plan.PlanDescriptionItems),
			&features,
		); err != nil {

			logger.Error(err.Error())

			return nil, status.Error(
				codes.Internal,
				"Invalid plan description items",
			)
		}
	}

	//----------------------------------------------------------------------
	// ELIMINAR FEATURE
	//----------------------------------------------------------------------

	filteredFeatures := make(
		[]map[string]json.RawMessage,
		0,
		len(features),
	)

	// for _, feature := range features {

	// 	var featureID int32

	// 	if rawID, exists := feature["id"]; exists {

	// 		if err := json.Unmarshal(rawID, &featureID); err != nil {
	// 			logger.Error(err.Error())

	// 			return nil, status.Error(
	// 				codes.Internal,
	// 				"Invalid feature ID",
	// 			)
	// 		}
	// 	}

	// 	if featureID != id {
	// 		filteredFeatures = append(
	// 			filteredFeatures,
	// 			feature,
	// 		)
	// 	}
	// }

	for _, feature := range features {
		rawID, exists := feature["id"]

		// Si la característica no tiene ID, conservarla.
		if !exists {
			filteredFeatures = append(filteredFeatures, feature)
			continue
		}

		// Un ID ausente o null no debe confundirse con el índice 0.
		var featureID *int32
		if err := json.Unmarshal(rawID, &featureID); err != nil {
			logger.Error(err.Error())

			return nil, status.Error(
				codes.Internal,
				"Invalid feature ID",
			)
		}

		if featureID == nil {
			filteredFeatures = append(filteredFeatures, feature)
			continue
		}

		// Eliminar únicamente la característica con el ID solicitado.
		if *featureID != id {
			filteredFeatures = append(filteredFeatures, feature)
		}
	}

	//----------------------------------------------------------------------
	// CONVERTIR ARRAY → JSON STRING
	//----------------------------------------------------------------------

	featuresJSON, err := json.Marshal(filteredFeatures)

	if err != nil {

		logger.Error(err.Error())

		return nil, status.Error(
			codes.Internal,
			"Internal server error",
		)
	}

	//----------------------------------------------------------------------
	// PERSISTENCIA
	//----------------------------------------------------------------------

	update := bson.M{
		"plan_description_items": string(featuresJSON),
		"updatedAt":              time.Now(),
	}

	if err := s.repository.UpdateByID(
		ctx,
		idP,
		update,
	); err != nil {

		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, status.Error(
				codes.NotFound,
				"Plan not found",
			)
		}

		logger.Error(err.Error())

		return nil, status.Error(
			codes.Internal,
			"Internal server error",
		)
	}

	//----------------------------------------------------------------------
	// OBTENER PLAN ACTUALIZADO
	//----------------------------------------------------------------------

	updatedPlan, err := s.repository.FindByID(
		ctx,
		idP,
	)

	if err != nil {

		logger.Error(err.Error())

		return nil, status.Error(
			codes.Internal,
			"Internal server error",
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
		pulsar.TopicPlanUpdated,
		"ms-plan",
		updatedPlan,
	)

	event.WithCorrelation(idP.Hex())

	if err := pulsar.Publish(
		ctx,
		pulsar.TopicPlanUpdated,
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
			"topic", pulsar.TopicPlanUpdated,
			"entity_id", idP.Hex(),
		)
	}

	return updatedPlan, nil
}
