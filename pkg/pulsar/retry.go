package pulsar

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// Este archivo maneja la tolerancia a fallos mediante reintentos con Exponential Backoff y Jitter
// (variación aleatoria) para no saturar los servicios en caso de caídas intermitentes.

// RetryConfig define la configuración para la estrategia de reintentos.
type RetryConfig struct {
	MaxRetries   int           // Número máximo de intentos
	InitialDelay time.Duration // Tiempo base de espera inicial
	MaxDelay     time.Duration // Tiempo máximo de espera
	Factor       float64       // Multiplicador exponencial
}

// DefaultRetryConfig entrega una configuración robusta por defecto.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:   3,
		InitialDelay: 200 * time.Millisecond,
		MaxDelay:     3 * time.Second,
		Factor:       2.0,
	}
}

// ExecuteWithRetry ejecuta un Handler envolviéndolo en una política de reintentos.
func ExecuteWithRetry(ctx context.Context, cfg RetryConfig, handler Handler, payload []byte) error {
	var err error

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		// Verificar si el contexto fue cancelado desde arriba (ej. shutdown)
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("retry canceled by context: %w", err)
		}

		err = handler(ctx, payload)
		if err == nil {
			return nil // Ejecución exitosa, salimos del ciclo
		}

		// Si es el último intento, no esperamos, simplemente rompemos
		if attempt == cfg.MaxRetries {
			break
		}

		// Cálculo de Exponential Backoff
		delay := float64(cfg.InitialDelay) * math.Pow(cfg.Factor, float64(attempt))
		if delay > float64(cfg.MaxDelay) {
			delay = float64(cfg.MaxDelay)
		}

		// Agregar Jitter (variación de +/- 20%) para evitar el efecto manada (thundering herd)
		jitter := (rand.Float64()*0.4 - 0.2) * delay
		sleepDuration := time.Duration(delay + jitter)

		select {
		case <-time.After(sleepDuration): // Esperamos antes del próximo reintento
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return fmt.Errorf("exceeded max retries (%d): %w", cfg.MaxRetries, err)
}
