package pulsar

import (
	"fmt"
	"sync"

	pulsarclient "github.com/apache/pulsar-client-go/pulsar"

	"github.com/SDATA-Colombia/ms-plan/internal/config"
)

var (
	client pulsarclient.Client
	once   sync.Once
)

// Init inicializa el cliente singleton de Apache Pulsar.
// Debe ejecutarse una sola vez durante el arranque del microservicio.
func Init(cfg *config.Config) error {

	var err error

	once.Do(func() {

		client, err = pulsarclient.NewClient(
			pulsarclient.ClientOptions{
				URL: cfg.PulsarURL,
			},
		)

	})

	if err != nil {
		return fmt.Errorf("failed to initialize Pulsar client: %w", err)
	}

	return nil
}

// Client retorna el cliente singleton de Apache Pulsar.
func Client() pulsarclient.Client {

	if client == nil {
		panic("pulsar client has not been initialized")
	}

	return client
}

// IsConnected indica si el cliente fue inicializado.
func IsConnected() bool {
	return client != nil
}
