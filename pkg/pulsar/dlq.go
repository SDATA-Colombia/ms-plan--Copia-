package pulsar

import (
	"context"
	"fmt"
	"time"
)

// Este archivo encapsula la lógica para enviar a la Dead Letter Queue los mensajes que fallaron definitivamente después de agotar los reintentos.

// DLQEnvelope empaqueta el evento original junto a la metadata de seguridad y error.
type DLQEnvelope struct {
	OriginalTopic string    `json:"original_topic"`
	ErrorReason   string    `json:"error_reason"`
	FailedAt      time.Time `json:"failed_at"`
	Payload       []byte    `json:"payload"`
}

// SendToDLQ redirige un payload irrecuperable a la cola de mensajes muertos.
func SendToDLQ(ctx context.Context, originalTopic string, payload []byte, reason error) error {
	dlqTopic := fmt.Sprintf("%s.dlq", originalTopic)

	envelope := DLQEnvelope{
		OriginalTopic: originalTopic,
		ErrorReason:   reason.Error(),
		FailedAt:      time.Now().UTC(),
		Payload:       payload,
	}

	// Reutilizamos el pool de Producers de producer.go
	if err := Publish(ctx, dlqTopic, envelope); err != nil {
		return fmt.Errorf("failed to route message to DLQ topic %s: %w", dlqTopic, err)
	}

	return nil
}
