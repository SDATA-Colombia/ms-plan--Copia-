//go:build defects

package mapper

// Pruebas que EXPONEN defectos conocidos de ms-plan.
// No forman parte de la suite normal porque hoy FALLAN a proposito.
// Ejecutar con:  go test -tags defects ./internal/mapper/...
// Cuando el defecto se corrige, la prueba pasa y se puede mover a mapper_test.go.

import (
	"testing"

	"github.com/SDATA-Colombia/ms-plan/internal/model/entity"
	planpb "github.com/SDATA-Colombia/shared-go-proto/plan/v1"
)

// D-01: EntityToProto(nil) provoca panic (nil pointer dereference).
// Camino real: service.UpdateByID retorna (nil, nil) si el plan desaparece entre
// la actualizacion y la lectura; el controller llama EntityToProto(updatedPlan).
func TestDefect_D01_EntityToProtoNilNoDebeHacerPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("D-01: EntityToProto(nil) hizo panic: %v", r)
		}
	}()

	_ = EntityToProto(nil)
}

// D-02: un PlanIDCompany vacio se serializa como "000000000000000000000000"
// en lugar de cadena vacia, lo que impide al cliente distinguir "sin empresa".
func TestDefect_D02_CompanyVaciaDebeSerializarseVacia(t *testing.T) {
	got := EntityToProto(&entity.Plan{})

	if got.GetPlanIdCompany() != "" {
		t.Errorf("D-02: PlanIdCompany vacio se serializo como %q", got.GetPlanIdCompany())
	}
	if got.GetId() != "" {
		t.Errorf("D-02: Id vacio se serializo como %q", got.GetId())
	}
}

// D-03: el round trip de una entidad con ObjectIDs en cero no es identidad:
// "" -> zero -> "0000..." -> zero. El proto de salida difiere del de entrada.
func TestDefect_D03_RoundTripConIDsVaciosEsIdentidad(t *testing.T) {
	in := &planpb.PlanEntity{PlanTitle: "X"}

	ent, err := ProtoToEntity(in)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	out := EntityToProto(ent)
	if out.GetId() != in.GetId() || out.GetPlanIdCompany() != in.GetPlanIdCompany() {
		t.Errorf("D-03: ids de entrada (%q,%q) != salida (%q,%q)",
			in.GetId(), in.GetPlanIdCompany(), out.GetId(), out.GetPlanIdCompany())
	}
}
