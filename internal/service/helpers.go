package service

import (
	"context"

	"github.com/SDATA-Colombia/ms-plan/pkg/cache"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// ----------------------------------------------------------------------
// HELPERS
// ----------------------------------------------------------------------

// Exists verifica si existe al menos una entidad según el filtro.
func (s *BaseService) Exists(
	ctx context.Context,
	filter bson.M,

) (bool, error) {

	return s.repository.Exists(
		ctx,
		filter,
	)
}

// Count retorna la cantidad de entidades que cumplen el filtro.
func (s *BaseService) Count(
	ctx context.Context,
	filter bson.M,

) (int64, error) {

	return s.repository.Count(
		ctx,
		filter,
	)
}

// bumpCacheVersion incrementa la versión del namespace.
//
// No elimina llaves de Dragonfly. Únicamente incrementa la versión
// del namespace para que las siguientes consultas utilicen nuevas
// llaves de caché. Las llaves anteriores expirarán automáticamente
// según su TTL.
func (s *BaseService) bumpCacheVersion(
	ctx context.Context,
	namespace ...any,

) error {

	parts := append([]any{cacheEntity}, namespace...)

	_, err := cache.IncrementVersion(
		ctx,
		cache.VersionKey(parts...),
	)

	return err
}

// Obtener la versión actual del namespace de caché.
func (s *BaseService) cacheVersion(
	ctx context.Context,
	namespace ...any,

) (int64, error) {

	parts := append([]any{cacheEntity}, namespace...)

	return cache.GetVersion(
		ctx,
		cache.VersionKey(parts...),
	)
}

// Construye una llave de caché versionada para el namespace de la entidad.
func (s *BaseService) cacheKey(
	ctx context.Context,
	tenantNamespace []any, // Ej: []any{"company", idCompany}
	parts ...any, // Ej: "one", id.Hex()

) (string, error) {

	// 1. Obtenemos la versión ESPECÍFICA de este tenant
	version, err := s.cacheVersion(ctx, tenantNamespace...)
	if err != nil {
		return "", err
	}

	// 2. Construimos la base: "cacheEntity:company:123:v:2"
	baseNamespace := append([]any{cacheEntity}, tenantNamespace...)

	// Aprovechamos tu función CacheKey de keys.go (Source 4)
	return cache.CacheKey(version, append(baseNamespace, parts...)...), nil
}
