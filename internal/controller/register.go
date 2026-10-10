package controller

// Controller: El traductor de gRPC.
// Recibe el request de proto, lo pasa al mapper, llama al service, devuelve proto.
// NO tiene lógica de negocio ni toca Mongo directamente.

//-------------------------------------------------------------------------------------------------------
// Node.js (Express + Mongoose): El envoltorio try/catch del controller de Express + res.status (...)
//-------------------------------------------------------------------------------------------------------

import (
	planv1 "github.com/SDATA-Colombia/shared-go-proto/plan/v1" // Asegúrate de importar la proto de ms-plan si es distinta

	"github.com/SDATA-Colombia/ms-plan/internal/service"

	"google.golang.org/grpc"
)

func Register(
	server *grpc.Server,
	baseService *service.BaseService,

) {
	// Plan
	planv1.RegisterPlanServiceServer(
		server,
		NewPlanController(baseService),
	)
}
