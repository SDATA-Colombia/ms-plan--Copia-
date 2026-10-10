package pulsar

import (
	"context"
	"time"

	"github.com/SDATA-Colombia/ms-plan/pkg/cache"
)

// HandlerIdempotent es un wrapper que asegura que un Handler se ejecute solo una vez por mensaje.
// Asegurando que los eventos no se procesen dos veces si Pulsar entrega un duplicado.

// IdempotencyStore define la interfaz necesaria para validar el procesamiento único.
type IdempotencyStore interface {
	IsProcessed(ctx context.Context, messageID string) (bool, error)
	MarkAsProcessed(ctx context.Context, messageID string, ttl time.Duration) error
}

// DragonflyIdempotencyStore implementación distribuida usando tu paquete cache.
type DragonflyIdempotencyStore struct {
	serviceName string
}

// NewDragonflyIdempotencyStore crea un nuevo store acoplado a Dragonfly.
func NewDragonflyIdempotencyStore(serviceName string) *DragonflyIdempotencyStore {
	return &DragonflyIdempotencyStore{
		serviceName: serviceName,
	}
}

// IsProcessed verifica en Dragonfly si el ID del mensaje ya existe.
func (d *DragonflyIdempotencyStore) IsProcessed(ctx context.Context, messageID string) (bool, error) {
	// Genera una llave unificada, ej: idempotency:sdata-ms-base:12345-67890
	key := cache.Key("idempotency", d.serviceName, messageID)

	return cache.Exists(ctx, key)
}

// MarkAsProcessed guarda el ID en Dragonfly con un TTL definido (ej. 24h).
func (d *DragonflyIdempotencyStore) MarkAsProcessed(ctx context.Context, messageID string, ttl time.Duration) error {
	key := cache.Key("idempotency", d.serviceName, messageID)

	// Guardamos un valor simple (true) para indicar que fue procesado con éxito.
	return cache.Set(ctx, key, true, ttl)
}
