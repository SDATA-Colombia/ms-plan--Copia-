package consumer

import (
	"context"

	"github.com/SDATA-Colombia/ms-plan/pkg/logger"
)

func PlanCreated(
	ctx context.Context,
	payload []byte,

) error {

	logger.Info(
		"Evento recibido",
		"payload",
		string(payload),
	)

	return nil
}
