package repository

// Repository: Solo habla con Mongo, nada más.
// Sabe insertar, buscar, contar, actualizar documentos, etc.
// NO sabe de reglas de negocio (ej: "no permitir emails duplicados", "enviar evento a Pulsar", "cachear en Dragonfly").
// NO sabe de gRPC ni de proto.
// Es "tonto a propósito" — solo traduce entre bson.M y Mongo.

//-------------------------------------------------------------------------------------------------------
// Node.js (Express + Mongoose): Model.create(), Model.findById() (Mongoose)
//-------------------------------------------------------------------------------------------------------

import (
	"context"

	"github.com/SDATA-Colombia/ms-plan/pkg/mongodb"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type BaseRepository[T any] struct {
	collection *mongo.Collection
}

func NewBaseRepository[T any](collection string) *BaseRepository[T] {
	return &BaseRepository[T]{
		collection: mongodb.Collection(collection),
	}
}

// Collection retorna la colección de MongoDB.
func (r *BaseRepository[T]) Collection() *mongo.Collection {
	return r.collection
}

// ----------------------------------------------------------------------
// INDEXES
// ----------------------------------------------------------------------

// EnsureIndexes crea los índices definidos para la colección.
// Si el índice ya existe, MongoDB simplemente lo reutiliza.
func (r *BaseRepository[T]) EnsureIndexes(
	ctx context.Context,
	indexModels []mongo.IndexModel,

) error {
	_, err := r.collection.Indexes().CreateMany(
		ctx,
		indexModels,
	)

	return err
}

// ----------------------------------------------------------------------
// CREATE
// ----------------------------------------------------------------------

// Create inserta una nueva entidad.
func (r *BaseRepository[T]) Create(
	ctx context.Context,
	entity *T,

) error {
	_, err := r.collection.InsertOne(
		ctx,
		entity,
	)

	return err
}

// ----------------------------------------------------------------------
// READ
// ----------------------------------------------------------------------

// FindByID obtiene una entidad por su identificador.
func (r *BaseRepository[T]) FindByID(
	ctx context.Context,
	id bson.ObjectID,

) (*T, error) {
	var entity T

	err := r.collection.FindOne(
		ctx,
		bson.M{
			"_id": id,
		},
	).Decode(&entity)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &entity, nil
}

// FindOne obtiene una entidad según el filtro.
func (r *BaseRepository[T]) FindOne(
	ctx context.Context,
	filter bson.M,

) (*T, error) {
	var entity T

	err := r.collection.FindOne(
		ctx,
		filter,
	).Decode(&entity)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &entity, nil
}

// FindManyPaginated obtiene múltiples entidades con paginación.
func (r *BaseRepository[T]) FindManyPaginated(
	ctx context.Context,
	filter bson.M,
	page int64,
	limit int64,

	// ) ([]*T, error) {
) (entities []*T, retErr error) {

	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 10
	}

	skip := (page - 1) * limit

	cursor, err := r.collection.Find(
		ctx,
		filter,
		options.Find().
			SetSkip(skip).
			SetLimit(limit),
	)

	if err != nil {
		return nil, err
	}

	// defer cursor.Close(ctx)

	// var entities []*T

	// if err := cursor.All(ctx, &entities); err != nil {
	// 	return nil, err
	// }

	// return entities, nil

	defer func() {
		if closeErr := cursor.Close(ctx); closeErr != nil && retErr == nil {
			retErr = closeErr
		}
	}()

	if err := cursor.All(ctx, &entities); err != nil {
		return nil, err
	}

	return entities, nil
}

// FindByField obtiene una entidad por un campo específico.
func (r *BaseRepository[T]) FindByField(
	ctx context.Context,
	field string,
	value any,

) (*T, error) {
	return r.FindOne(
		ctx,
		bson.M{
			field: value,
		},
	)
}

// ----------------------------------------------------------------------
// HELPERS
// ----------------------------------------------------------------------

// Exists verifica si existe al menos una entidad según el filtro.
func (r *BaseRepository[T]) Exists(
	ctx context.Context,
	filter bson.M,

) (bool, error) {
	count, err := r.collection.CountDocuments(
		ctx,
		filter,
	)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Count retorna la cantidad de entidades que cumplen el filtro.
func (r *BaseRepository[T]) Count(
	ctx context.Context,
	filter bson.M,

) (int64, error) {
	return r.collection.CountDocuments(
		ctx,
		filter,
	)
}

// ----------------------------------------------------------------------
// UPDATE
// ----------------------------------------------------------------------

// UpdateByID actualiza una entidad por su identificador.
func (r *BaseRepository[T]) UpdateByID(
	ctx context.Context,
	id bson.ObjectID,
	update bson.M,

) error {
	updateDoc := bson.D{
		{
			Key:   "$set",
			Value: update,
		},
	}

	result, err := r.collection.UpdateOne(
		ctx,
		bson.D{
			{
				Key:   "_id",
				Value: id,
			},
		},
		updateDoc,
	)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

// UpdateMany actualiza múltiples entidades según el filtro.
func (r *BaseRepository[T]) UpdateMany(
	ctx context.Context,
	filter bson.M,
	update bson.M,

) error {
	_, err := r.collection.UpdateMany(
		ctx,
		filter,
		bson.M{
			"$set": update,
		},
	)

	return err
}

// ----------------------------------------------------------------------
// DELETE
// ----------------------------------------------------------------------

// DeleteByID elimina físicamente una entidad por su identificador.
func (r *BaseRepository[T]) DeleteByID(
	ctx context.Context,
	id bson.ObjectID,

) error {
	result, err := r.collection.DeleteOne(
		ctx,
		bson.M{
			"_id": id,
		},
	)

	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}
