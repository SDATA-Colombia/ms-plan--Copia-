package pulsar

import (
	"sync"

	pulsarclient "github.com/apache/pulsar-client-go/pulsar"
)

// Close libera todos los recursos utilizados por Apache Pulsar.
func Close() {

	//---------------------------------------------------
	// Cerrar Consumers
	//---------------------------------------------------

	mutex.Lock()

	for _, consumer := range consumers {

		if consumer != nil {
			consumer.Close()
		}
	}

	consumers = nil

	mutex.Unlock()

	//---------------------------------------------------
	// Cerrar Producers
	//---------------------------------------------------

	producerMutex.Lock()

	for topic, producer := range producers {

		if producer != nil {
			producer.Close()
		}

		delete(producers, topic)
	}

	producerMutex.Unlock()

	//---------------------------------------------------
	// Cerrar Client
	//---------------------------------------------------

	if client != nil {
		client.Close()
		client = nil
	}
}

// Reset elimina el estado interno.
// Útil para pruebas unitarias.
func Reset() {

	mutex.Lock()
	consumers = nil
	mutex.Unlock()

	producerMutex.Lock()
	producers = make(map[string]pulsarclient.Producer)
	producerMutex.Unlock()

	client = nil

	once = sync.Once{}
}
