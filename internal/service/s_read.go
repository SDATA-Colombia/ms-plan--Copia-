package service

import (
	"context"

	"github.com/SDATA-Colombia/ms-plan/internal/model/entity"
	"github.com/SDATA-Colombia/ms-plan/pkg/cache"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// BaseService está en el mismo package service.
// repository ya está importado en base.go.
// T es el genérico declarado en BaseService[T].
// repository.FindByID(), etc; ya son accesibles a través de s.repository.
// En Go cada archivo importa únicamente lo que utiliza, aunque pertenezcan al mismo package.

// ----------------------------------------------------------------------
// READ
// ----------------------------------------------------------------------

// FindByID obtiene una entidad por su identificador.
func (s *BaseService) FindByID(
	ctx context.Context,
	id bson.ObjectID,

) (*entity.Plan, error) {

	// Dragonfly: Generar llave de caché.
	// Si no tienes el company_id aquí, usamos un namespace global o lo dejamos vacío.
	// Lo ideal sería obtenerlo del ctx: companyID := ctx.Value("company_id")
	key, err := s.cacheKey(ctx, []any{"plan", "global"}, "one", id.Hex())
	if err != nil {
		return nil, err
	}

	var result entity.Plan

	found, err := cache.Get(ctx, key, &result)
	if err != nil {
		return nil, err
	}

	if found {
		return &result, nil
	}

	// No estaba en caché, buscar en MongoDB.
	resultPtr, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Guardar en caché el resultado.
	if resultPtr != nil {
		_ = cache.Set(ctx, key, resultPtr, cache.TTLStandard)
	}

	return resultPtr, nil
}

// FindManyPaginated obtiene múltiples entidades paginadas.
func (s *BaseService) FindManyPaginated(
	ctx context.Context,
	filter bson.M,
	page int64,
	limit int64,

) ([]*entity.Plan, error) {

	// Dragonfly: Generar llave de caché.
	hash, err := cache.Hash([]any{
		filter,
		page,
		limit,
	})
	if err != nil {
		return nil, err
	}

	var tenantID any = "global"

	key, err := s.cacheKey(ctx, []any{"plan", tenantID}, "page", hash)
	if err != nil {
		return nil, err
	}

	var result []*entity.Plan

	found, err := cache.Get(ctx, key, &result)
	if err != nil {
		return nil, err
	}

	if found {
		return result, nil
	}

	// No estaba en caché, buscar en MongoDB.
	result, err = s.repository.FindManyPaginated(
		ctx,
		filter,
		page,
		limit,
	)
	if err != nil {
		return nil, err
	}

	// Guardar en caché el resultado.
	_ = cache.Set(ctx, key, result, cache.TTLRealtime)

	return result, nil
}

// FindByField obtiene una entidad por un campo.
func (s *BaseService) FindByField(
	ctx context.Context,
	field string,
	value any,

) (*entity.Plan, error) {

	// Como no tenemos un 'filter', definimos el tenant dinámicamente.
	// Si la búsqueda casualmente es por empresa, usamos el value.
	// Si no, usamos "global" o lo sacas del ctx.
	var tenantNamespace []any
	if field == "company_id" {
		tenantNamespace = []any{"plan", value} // <-- Corregido: "company" -> "plan"
	} else {
		tenantNamespace = []any{"plan", "global"} // <-- Corregido: Agregado "plan"
	}

	// Dragonfly: Generar llave de caché.
	key, err := s.cacheKey(ctx, tenantNamespace, "field", field, value)
	if err != nil {
		return nil, err
	}

	var result entity.Plan

	found, err := cache.Get(ctx, key, &result)
	if err != nil {
		return nil, err
	}

	if found {
		return &result, nil
	}

	// No estaba en caché, buscar en MongoDB.
	resultPtr, err := s.repository.FindByField(
		ctx,
		field,
		value,
	)
	if err != nil {
		return nil, err
	}

	// Guardar en caché el resultado.
	if resultPtr != nil {
		_ = cache.Set(ctx, key, resultPtr, cache.TTLStandard)
	}

	return resultPtr, nil
}