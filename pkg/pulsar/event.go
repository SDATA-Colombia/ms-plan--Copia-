package pulsar

import (
	"time"

	"github.com/google/uuid"
)

// Event representa un evento publicado en Apache Pulsar.
// Es la estructura base utilizada por todos los microservicios.
type Event struct {
	ID            string    `json:"id"`
	Type          string    `json:"type"`
	Source        string    `json:"source"`
	Version       string    `json:"version"`
	OccurredAt    time.Time `json:"occurred_at"`
	CorrelationID string    `json:"correlation_id,omitempty"`
	UserID        string    `json:"user_id,omitempty"`
	TenantID      string    `json:"tenant_id,omitempty"`
	Payload       any       `json:"payload"`
}

// NewEvent crea un nuevo evento con valores por defecto.
func NewEvent(
	eventType string,
	source string,
	payload any,

) Event {

	if payload == nil {
		payload = struct{}{}
	}

	return Event{
		ID:         uuid.NewString(),
		Type:       eventType,
		Source:     source,
		Version:    "1.0.0",
		OccurredAt: time.Now().UTC(),
		Payload:    payload,
	}
}

// WithCorrelation establece el Correlation ID.
func (e *Event) WithCorrelation(correlationID string) {
	e.CorrelationID = correlationID
}

// WithUser establece el usuario que originó el evento.
func (e *Event) WithUser(userID string) {
	e.UserID = userID
}

// WithTenant establece el Tenant del evento.
func (e *Event) WithTenant(tenantID string) {
	e.TenantID = tenantID
}
