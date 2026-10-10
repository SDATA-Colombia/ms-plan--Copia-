package consumer

import "github.com/SDATA-Colombia/ms-plan/pkg/pulsar"

func Register() error {

	// 1. Instanciamos la configuración de reintentos por defecto
	retryCfg := pulsar.DefaultRetryConfig()

	// 2. Instanciamos el almacén de idempotencia usando tu caché en Dragonfly
	store := pulsar.NewDragonflyIdempotencyStore("ms-plan")

	return pulsar.Subscribe(
		pulsar.TopicPlanCreated,
		"ms-plan",
		PlanCreated,
		retryCfg, // Inyectamos Retry
		store,    // Inyectamos Idempotencia
	)
}
