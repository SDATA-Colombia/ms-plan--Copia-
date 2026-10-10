package pulsar

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	pulsarclient "github.com/apache/pulsar-client-go/pulsar"
)

var (
	producers     = make(map[string]pulsarclient.Producer)
	producerMutex sync.RWMutex
)

// Producer retorna (o crea) un Producer para un tópico.
// Cada tópico mantiene un único Producer reutilizable.
func Producer(topic string) (pulsarclient.Producer, error) {

	// Buscar si ya existe.
	producerMutex.RLock()

	producer, ok := producers[topic]

	producerMutex.RUnlock()

	if ok {
		return producer, nil
	}

	// Crear uno nuevo.
	producerMutex.Lock()
	defer producerMutex.Unlock()

	// Doble verificación por concurrencia.
	if producer, ok := producers[topic]; ok {
		return producer, nil
	}

	var err error

	producer, err = Client().CreateProducer(
		pulsarclient.ProducerOptions{
			Topic: topic,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	producers[topic] = producer

	return producer, nil
}

// Publish serializa un evento y lo publica en el tópico indicado.
func Publish(
	ctx context.Context,
	topic string,
	event any,

) error {

	producer, err := Producer(topic)
	if err != nil {
		return err
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Evita que una publicación quede bloqueada indefinidamente.
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err = producer.Send(
		ctx,
		&pulsarclient.ProducerMessage{
			Payload: data,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}
