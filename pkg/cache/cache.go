package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/SDATA-Colombia/ms-plan/pkg/dragonfly"
	"github.com/redis/go-redis/v9"
)

// Get obtiene un valor desde Dragonfly y lo deserializa.
func Get(
	ctx context.Context,
	key string,
	dest any,

) (bool, error) {

	data, err := dragonfly.Client().
		Get(ctx, key).
		Result()

	if err == redis.Nil {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return false, err
	}

	return true, nil
}

// Set almacena un valor serializado en Dragonfly.
func Set(
	ctx context.Context,
	key string,
	value any,
	ttl time.Duration,

) error {

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return dragonfly.Client().
		Set(ctx, key, data, ttl).
		Err()
}

// Delete elimina una o varias llaves.
func Delete(
	ctx context.Context,
	keys ...string,

) error {

	return dragonfly.Client().
		Del(ctx, keys...).
		Err()
}

// Exists indica si una llave existe.
func Exists(
	ctx context.Context,
	key string,

) (bool, error) {

	n, err := dragonfly.Client().
		Exists(ctx, key).
		Result()

	if err != nil {
		return false, err
	}

	return n > 0, nil
}
