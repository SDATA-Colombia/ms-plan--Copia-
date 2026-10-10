package controller

import (
	"github.com/SDATA-Colombia/ms-plan/internal/service"

	v1 "github.com/SDATA-Colombia/shared-go-proto/plan/v1"
)

// gRPC: PlanController implementa la interfaz gRPC generada por protoc-gen-go-grpc.
type PlanController struct {
	v1.UnimplementedPlanServiceServer
	service *service.BaseService
}

func NewPlanController(
	service *service.BaseService,
	
) *PlanController {

	return &PlanController{
		service: service,
	}
}

// Funciones del PlanController que implementan la interfaz gRPC generada por protoc-gen-go-grpc.
// Están implementadas en archivos separados con contexto por cada función CRUD.
// Ruta: /controller/create.go, /controller/delete.go, etc.
// Cada función recibe un contexto, una solicitud y retorna una respuesta y un error.
