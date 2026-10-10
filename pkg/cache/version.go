package cache

import (
	"context"
	"strconv"

	"github.com/SDATA-Colombia/ms-plan/pkg/dragonfly"
	"github.com/redis/go-redis/v9"
)

// GetVersion obtiene la versión actual de un namespace.
//
// Si la versión no existe, se inicializa automáticamente en 1.
//
// Ejemplo:
//
//	namespace := cache.VersionKey(
//	    "color",
//	    "company",
//	    companyID,
//	)
//
//	version, err := cache.GetVersion(ctx, namespace)
func GetVersion(
	ctx context.Context,
	namespace string,

) (int64, error) {

	client := dragonfly.Client()

	value, err := client.Get(ctx, namespace).Result()

	if err == redis.Nil {

		if _, err := client.SetNX(
			ctx,
			namespace,
			1,
			TTLForever,
		).Result(); err != nil {
			return 0, err
		}

		value, err = client.Get(ctx, namespace).Result()
		if err != nil {
			return 0, err
		}
	}

	if err != nil {
		return 0, err
	}

	version, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, err
	}

	return version, nil
}

// IncrementVersion incrementa la versión de un namespace.
//
// Todas las llaves construidas con la versión anterior
// quedarán automáticamente obsoletas.
func IncrementVersion(
	ctx context.Context,
	namespace string,

) (int64, error) {

	return dragonfly.Client().
		Incr(ctx, namespace).
		Result()
}

// SetVersion establece manualmente una versión.
//
// Normalmente no se utiliza, pero resulta útil
// para migraciones o pruebas.
func SetVersion(
	ctx context.Context,
	namespace string,
	version int64,

) error {

	return dragonfly.Client().
		Set(
			ctx,
			namespace,
			version,
			TTLForever,
		).
		Err()
}

// DeleteVersion elimina una versión.
//
// La siguiente llamada a GetVersion()
// volverá a crearla con valor 1.
func DeleteVersion(
	ctx context.Context,
	namespace string,

) error {

	return dragonfly.Client().
		Del(ctx, namespace).
		Err()
}
