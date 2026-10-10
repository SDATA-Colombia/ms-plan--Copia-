package service

// Service: Capa de negocio / lógica de negocio.
// Decide cuándo llamar al repository.
// Aquí van las validaciones, el cache-aside con Dragonfly,
// publicar eventos a Pulsar, orquestar varias operaciones.
// NO sabe de gRPC ni de proto — solo trabaja con entity.Plan.

import (
	"time"

	"github.com/SDATA-Colombia/ms-plan/internal/model/entity"
	"github.com/SDATA-Colombia/ms-plan/internal/repository"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type BaseService struct {
	repository *repository.BaseRepository[entity.Plan]
}

func NewBaseService(
	repository *repository.BaseRepository[entity.Plan],

) *BaseService {

	return &BaseService{
		repository: repository,
	}
}

// ----------------------------------------------------------------------
// DEFAULTS
// ----------------------------------------------------------------------

// applyCreateDefaults aplica los valores por defecto para una creación.
func (s *BaseService) applyCreateDefaults(
	entity *entity.Plan,
) {

	if entity == nil {
		return
	}

	now := time.Now().UTC()

	// MongoDB genera automáticamente el _id,
	// pero si queremos comportarnos como Mongoose
	// lo generamos antes de guardar.
	if entity.ID.IsZero() {
		entity.ID = bson.NewObjectID()
	}

	if entity.CreatedAt.IsZero() {
		entity.CreatedAt = now
	}

	entity.UpdatedAt = now
	entity.DeletedAt = nil

	// Valor por defecto.
	entity.PlanTotalSales = 0
}

// applyUpdateDefaults aplica los valores por defecto
// antes de actualizar un documento.
func (s *BaseService) applyUpdateDefaults(
	update bson.M,
) {

	if update == nil {
		return
	}

	update["updatedAt"] = time.Now().UTC()

	delete(update, "_id")
	delete(update, "createdAt")
	delete(update, "deletedAt")
}
