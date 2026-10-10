package pulsar

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/SDATA-Colombia/ms-plan/pkg/cache"

	pulsarclient "github.com/apache/pulsar-client-go/pulsar"
)

// Handler representa la función que procesará un mensaje recibido.
type Handler func(context.Context, []byte) error

var (
	consumers []pulsarclient.Consumer
	mutex     sync.RWMutex
)

// Subscribe crea un Consumer para un tópico y comienza
// a escuchar mensajes en una goroutine.
func Subscribe(
	topic string,
	subscription string,
	handler Handler,
	retryCfg RetryConfig,
	store IdempotencyStore,

) error {

	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	consumer, err := Client().Subscribe(
		pulsarclient.ConsumerOptions{
			Topic:            topic,
			SubscriptionName: subscription,
			Type:             pulsarclient.Shared,
		},
	)

	if err != nil {
		return fmt.Errorf("error creating consumer: %w", err)
	}

	mutex.Lock()
	consumers = append(consumers, consumer)
	mutex.Unlock()

	go consume(consumer, topic, handler, retryCfg, store)

	return nil
}

// Consumers retorna todos los consumers registrados.
func Consumers() []pulsarclient.Consumer {

	mutex.RLock()
	defer mutex.RUnlock()

	result := make([]pulsarclient.Consumer, len(consumers))
	copy(result, consumers)

	return result
}

// consume mantiene un loop infinito procesando mensajes.
func consume(
	consumer pulsarclient.Consumer,
	topic string,
	handler Handler,
	retryCfg RetryConfig,
	store IdempotencyStore,
) {

	ctx := context.Background()

	for {

		msg, err := consumer.Receive(ctx)
		// if err != nil {
		// 	return
		// }

		if err != nil {
			log.Printf("pulsar: receive failed for topic %q: %v", topic, err)
			return
		}

		//------------------------------------------------------------
		// Deserializar el evento para obtener el Event.ID
		//------------------------------------------------------------

		var event Event

		// if err := json.Unmarshal(msg.Payload(), &event); err != nil {
		// 	_ = SendToDLQ(ctx, topic, msg.Payload(), err)
		// 	consumer.Ack(msg)
		// 	continue
		// }

		if err := json.Unmarshal(msg.Payload(), &event); err != nil {
			if dlqErr := SendToDLQ(ctx, topic, msg.Payload(), err); dlqErr != nil {
				log.Printf("pulsar: DLQ publish failed for topic %q: %v", topic, dlqErr)
				consumer.Nack(msg)
				continue
			}

			ackOrRequeue(consumer, msg, topic)
			continue
		}

		//------------------------------------------------------------
		// 1. Idempotencia usando Event.ID
		//------------------------------------------------------------

		if store != nil {

			processed, err := store.IsProcessed(ctx, event.ID)

			// if err == nil && processed {
			// 	consumer.Ack(msg)
			// 	continue
			// }

			if err != nil {
				log.Printf("pulsar: idempotency check failed for topic %q: %v", topic, err)
				consumer.Nack(msg)
				continue
			}

			if processed {
				ackOrRequeue(consumer, msg, topic)
				continue
			}
		}

		//------------------------------------------------------------
		// 2. Retry
		//------------------------------------------------------------

		err = ExecuteWithRetry(ctx, retryCfg, handler, msg.Payload())

		if err != nil {

			//--------------------------------------------------------
			// 3. DLQ
			//--------------------------------------------------------

			// _ = SendToDLQ(ctx, topic, msg.Payload(), err)

			// consumer.Ack(msg)

			// continue

			if dlqErr := SendToDLQ(ctx, topic, msg.Payload(), err); dlqErr != nil {
				log.Printf("pulsar: DLQ publish failed for topic %q: %v", topic, dlqErr)
				consumer.Nack(msg)
				continue
			}

			ackOrRequeue(consumer, msg, topic)
			continue
		}

		//------------------------------------------------------------
		// 4. Marcar como procesado
		//------------------------------------------------------------

		if store != nil {
			// _ = store.MarkAsProcessed(ctx, event.ID, cache.TTLDaily)

			if err := store.MarkAsProcessed(ctx, event.ID, cache.TTLDaily); err != nil {
				log.Printf("pulsar: marking event %q as processed failed: %v", event.ID, err)
				consumer.Nack(msg)
				continue
			}
		}

		//------------------------------------------------------------
		// 5. ACK
		//------------------------------------------------------------

		// consumer.Ack(msg)

		// Confirmar el mensaje.
		ackOrRequeue(consumer, msg, topic)
	}
}

// ackOrRequeue confirma el mensaje y solicita su reentrega
// si falla la confirmación.
func ackOrRequeue(
	consumer pulsarclient.Consumer,
	msg pulsarclient.Message,
	topic string,
) {
	if err := consumer.Ack(msg); err != nil {
		log.Printf("pulsar: ACK failed for topic %q: %v", topic, err)
		consumer.Nack(msg)
	}
}
