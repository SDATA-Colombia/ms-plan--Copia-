//go:build integration

// Pruebas de INTEGRACION del CRUD de ms-plan.
//
// Levantan el servicio real (controller + service + repository) sobre un servidor
// gRPC en memoria (bufconn) y lo conectan a MongoDB, Dragonfly y Pulsar REALES,
// inicializados igual que en cmd/start.go.
//
// ADVERTENCIA: escriben en la base de datos configurada. Usa una base de datos de
// PRUEBAS (variables de entorno de internal/config), nunca la de produccion.
//
// Ejecucion:
//
//	MSPLAN_IT=1 go test -tags integration -v ./tests/integration/...
//	(agregar MSPLAN_IT_DEFECTS=1 para correr las pruebas que exponen defectos)
//	(agregar MSPLAN_IT_ALLOW_FEATURED=1 para la prueba de plan destacado: modifica
//	 TODOS los planes con plan_featured=true de la base configurada)
package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/SDATA-Colombia/ms-plan/indexes"
	"github.com/SDATA-Colombia/ms-plan/internal/config"
	"github.com/SDATA-Colombia/ms-plan/internal/controller"
	"github.com/SDATA-Colombia/ms-plan/internal/model/entity"
	"github.com/SDATA-Colombia/ms-plan/internal/repository"
	"github.com/SDATA-Colombia/ms-plan/internal/service"
	"github.com/SDATA-Colombia/ms-plan/pkg/dragonfly"
	"github.com/SDATA-Colombia/ms-plan/pkg/logger"
	"github.com/SDATA-Colombia/ms-plan/pkg/mongodb"
	"github.com/SDATA-Colombia/ms-plan/pkg/pulsar"
	planpb "github.com/SDATA-Colombia/shared-go-proto/plan/v1"

	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

var client planpb.PlanServiceClient

// ---------------------------------------------------------------------
// Arranque (equivalente a cmd/start.go, sin escuchar en un puerto real)
// ---------------------------------------------------------------------

func TestMain(m *testing.M) {
	if os.Getenv("MSPLAN_IT") != "1" {
		fmt.Println("Pruebas de integracion omitidas: defina MSPLAN_IT=1 para ejecutarlas")
		os.Exit(0)
	}

	cfg := config.Load()
	logger.Init(cfg.Environment)

	if err := mongodb.Init(cfg); err != nil {
		fmt.Println("MongoDB:", err)
		os.Exit(1)
	}
	if err := indexes.Init(context.Background()); err != nil {
		fmt.Println("indices:", err)
		os.Exit(1)
	}
	if err := dragonfly.Init(cfg); err != nil {
		fmt.Println("Dragonfly:", err)
		os.Exit(1)
	}
	if err := pulsar.Init(cfg); err != nil {
		if os.Getenv("MSPLAN_IT_NO_PULSAR") != "1" {
			fmt.Println("Pulsar:", err, "(use MSPLAN_IT_NO_PULSAR=1 para continuar sin Pulsar)")
			os.Exit(1)
		}
		fmt.Println("AVISO: Pulsar no disponible, los eventos no se publicaran:", err)
	}

	svc := service.NewBaseService(repository.NewBaseRepository[entity.Plan]("plans"))

	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer()
	controller.Register(srv, svc)
	go func() { _ = srv.Serve(lis) }()

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return lis.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		fmt.Println("cliente gRPC:", err)
		os.Exit(1)
	}
	client = planpb.NewPlanServiceClient(conn)

	code := m.Run()

	_ = conn.Close()
	srv.Stop()
	pulsar.Close()
	mongodb.Close()
	dragonfly.Close()

	os.Exit(code)
}

// ---------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------

func ctx10s(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func uniqueTitle(t *testing.T) string {
	t.Helper()
	name := strings.ReplaceAll(t.Name(), "/", "-")
	return "IT-" + name + "-" + bson.NewObjectID().Hex()[:8]
}

func newPlan(title string) *planpb.PlanEntity {
	return &planpb.PlanEntity{
		PlanTitle:         title,
		PlanSubtitle:      "Subtitulo " + title,
		PlanType:          "monthly",
		PlanCurrencyPrice: "COP",
		PlanDays:          30,
		PlanOffices:       1,
		PlanUsers:         5,
		PlanPrice:         99900,
		ActivePlan:        true,
	}
}

// createPlan crea un plan y registra su borrado al terminar la prueba.
func createPlan(t *testing.T, p *planpb.PlanEntity) *planpb.PlanEntity {
	t.Helper()

	res, err := client.CreatePlan(ctx10s(t), &planpb.CreatePlanReq{Data: p})
	if err != nil {
		t.Fatalf("CreatePlan fallo: %v", err)
	}

	created := res.GetData()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = client.DeletePlan(ctx, &planpb.DeletePlanReq{Id: created.GetId()})
	})

	return created
}

func wantCode(t *testing.T, err error, want codes.Code) {
	t.Helper()

	if err == nil {
		t.Fatalf("se esperaba codigo %s y no hubo error", want)
	}
	if got := status.Code(err); got != want {
		t.Fatalf("codigo: got %s want %s (err=%v)", got, want, err)
	}
}

func randomID() string { return bson.NewObjectID().Hex() }

func skipUnless(t *testing.T, envVar string) {
	t.Helper()
	if os.Getenv(envVar) != "1" {
		t.Skipf("omitida: defina %s=1", envVar)
	}
}

// ---------------------------------------------------------------------
// CREATE
// ---------------------------------------------------------------------

func TestCreatePlan_OK(t *testing.T) {
	in := newPlan(uniqueTitle(t))
	in.PlanTotalSales = 100 // el servicio debe forzar 0

	res, err := client.CreatePlan(ctx10s(t), &planpb.CreatePlanReq{Data: in})
	if err != nil {
		t.Fatalf("CreatePlan: %v", err)
	}
	created := res.GetData()
	t.Cleanup(func() {
		_, _ = client.DeletePlan(context.Background(), &planpb.DeletePlanReq{Id: created.GetId()})
	})

	if !res.GetSuccess() {
		t.Errorf("success debe ser true")
	}
	if _, err := bson.ObjectIDFromHex(created.GetId()); err != nil {
		t.Errorf("el id generado no es un ObjectID valido: %q", created.GetId())
	}
	if created.GetPlanTitle() != in.GetPlanTitle() {
		t.Errorf("titulo: got %q want %q", created.GetPlanTitle(), in.GetPlanTitle())
	}
	if created.GetPlanTotalSales() != 0 {
		t.Errorf("PlanTotalSales debe forzarse a 0, got %d", created.GetPlanTotalSales())
	}
	if created.GetCreatedAt() == nil || created.GetCreatedAt().AsTime().IsZero() {
		t.Errorf("CreatedAt debe estar definido")
	}
	if created.GetDeletedAt() != nil {
		t.Errorf("DeletedAt debe ser nil en un plan nuevo")
	}
}

func TestCreatePlan_MissingTitle(t *testing.T) {
	_, err := client.CreatePlan(ctx10s(t), &planpb.CreatePlanReq{
		Data: &planpb.PlanEntity{PlanSubtitle: "sin titulo"},
	})
	wantCode(t, err, codes.InvalidArgument)
	if !strings.Contains(err.Error(), "Plan title is required") {
		t.Errorf("mensaje inesperado: %v", err)
	}
}

func TestCreatePlan_MissingSubtitle(t *testing.T) {
	_, err := client.CreatePlan(ctx10s(t), &planpb.CreatePlanReq{
		Data: &planpb.PlanEntity{PlanTitle: uniqueTitle(t)},
	})
	wantCode(t, err, codes.InvalidArgument)
	if !strings.Contains(err.Error(), "Plan subtitle is required") {
		t.Errorf("mensaje inesperado: %v", err)
	}
}

func TestCreatePlan_NilData(t *testing.T) {
	_, err := client.CreatePlan(ctx10s(t), &planpb.CreatePlanReq{})
	wantCode(t, err, codes.InvalidArgument)
}

func TestCreatePlan_DuplicateTitle(t *testing.T) {
	title := uniqueTitle(t)
	createPlan(t, newPlan(title))

	_, err := client.CreatePlan(ctx10s(t), &planpb.CreatePlanReq{Data: newPlan(title)})
	wantCode(t, err, codes.AlreadyExists)
}

// ---------------------------------------------------------------------
// READ
// ---------------------------------------------------------------------

func TestGetPlan_OK(t *testing.T) {
	created := createPlan(t, newPlan(uniqueTitle(t)))

	res, err := client.GetPlan(ctx10s(t), &planpb.GetPlanReq{Id: created.GetId()})
	if err != nil {
		t.Fatalf("GetPlan: %v", err)
	}
	if res.GetData().GetId() != created.GetId() {
		t.Errorf("id: got %s want %s", res.GetData().GetId(), created.GetId())
	}
	if res.GetData().GetPlanTitle() != created.GetPlanTitle() {
		t.Errorf("titulo: got %q want %q", res.GetData().GetPlanTitle(), created.GetPlanTitle())
	}
}

// Segunda lectura: debe responder igual (camino de cache de Dragonfly).
func TestGetPlan_SecondReadSameResult(t *testing.T) {
	created := createPlan(t, newPlan(uniqueTitle(t)))

	first, err := client.GetPlan(ctx10s(t), &planpb.GetPlanReq{Id: created.GetId()})
	if err != nil {
		t.Fatalf("primera lectura: %v", err)
	}
	second, err := client.GetPlan(ctx10s(t), &planpb.GetPlanReq{Id: created.GetId()})
	if err != nil {
		t.Fatalf("segunda lectura (cache): %v", err)
	}
	if first.GetData().GetPlanTitle() != second.GetData().GetPlanTitle() ||
		first.GetData().GetPlanPrice() != second.GetData().GetPlanPrice() {
		t.Errorf("la lectura cacheada difiere de la lectura a MongoDB")
	}
}

func TestGetPlan_NotFound(t *testing.T) {
	_, err := client.GetPlan(ctx10s(t), &planpb.GetPlanReq{Id: randomID()})
	wantCode(t, err, codes.NotFound)
}

func TestGetPlan_InvalidID(t *testing.T) {
	_, err := client.GetPlan(ctx10s(t), &planpb.GetPlanReq{Id: "xyz"})
	wantCode(t, err, codes.InvalidArgument)
}

func TestListPlan_Pagination(t *testing.T) {
	for i := 0; i < 3; i++ {
		createPlan(t, newPlan(fmt.Sprintf("%s-%d", uniqueTitle(t), i)))
	}

	p1, err := client.ListPlan(ctx10s(t), &planpb.ListPlanReq{Page: 1, Limit: 2})
	if err != nil {
		t.Fatalf("ListPlan pagina 1: %v", err)
	}
	if len(p1.GetData()) != 2 {
		t.Errorf("pagina 1 con limit 2: got %d elementos", len(p1.GetData()))
	}
	if p1.GetTotal() < 3 {
		t.Errorf("total debe ser >= 3, got %d", p1.GetTotal())
	}

	p2, err := client.ListPlan(ctx10s(t), &planpb.ListPlanReq{Page: 2, Limit: 2})
	if err != nil {
		t.Fatalf("ListPlan pagina 2: %v", err)
	}
	if len(p2.GetData()) == 0 {
		t.Fatalf("pagina 2 no debe estar vacia con >= 3 planes")
	}
	for _, a := range p1.GetData() {
		for _, b := range p2.GetData() {
			if a.GetId() == b.GetId() {
				t.Errorf("el plan %s aparece en las paginas 1 y 2", a.GetId())
			}
		}
	}
}

func TestListPlan_DefaultsWhenZero(t *testing.T) {
	createPlan(t, newPlan(uniqueTitle(t)))

	res, err := client.ListPlan(ctx10s(t), &planpb.ListPlanReq{})
	if err != nil {
		t.Fatalf("ListPlan sin parametros: %v", err)
	}
	if !res.GetSuccess() {
		t.Errorf("success debe ser true")
	}
	if len(res.GetData()) == 0 || len(res.GetData()) > 10 {
		t.Errorf("por defecto debe devolver entre 1 y 10 elementos, got %d", len(res.GetData()))
	}
}

// ---------------------------------------------------------------------
// UPDATE
// ---------------------------------------------------------------------

func TestUpdatePlan_OK_InvalidatesCache(t *testing.T) {
	created := createPlan(t, newPlan(uniqueTitle(t)))

	// Poblar cache.
	if _, err := client.GetPlan(ctx10s(t), &planpb.GetPlanReq{Id: created.GetId()}); err != nil {
		t.Fatalf("GetPlan previo: %v", err)
	}

	upd := newPlan(created.GetPlanTitle())
	upd.PlanSubtitle = "Subtitulo actualizado"
	upd.PlanPrice = 150000

	res, err := client.UpdatePlan(ctx10s(t), &planpb.UpdatePlanReq{Id: created.GetId(), Data: upd})
	if err != nil {
		t.Fatalf("UpdatePlan: %v", err)
	}
	if res.GetData().GetPlanSubtitle() != "Subtitulo actualizado" || res.GetData().GetPlanPrice() != 150000 {
		t.Errorf("la respuesta no refleja la actualizacion: %+v", res.GetData())
	}

	// La lectura posterior NO debe devolver el valor viejo cacheado.
	got, err := client.GetPlan(ctx10s(t), &planpb.GetPlanReq{Id: created.GetId()})
	if err != nil {
		t.Fatalf("GetPlan posterior: %v", err)
	}
	if got.GetData().GetPlanSubtitle() != "Subtitulo actualizado" {
		t.Errorf("cache no invalidada: subtitulo=%q", got.GetData().GetPlanSubtitle())
	}
}

func TestUpdatePlan_NotFound(t *testing.T) {
	_, err := client.UpdatePlan(ctx10s(t), &planpb.UpdatePlanReq{
		Id:   randomID(),
		Data: newPlan(uniqueTitle(t)),
	})
	wantCode(t, err, codes.NotFound)
}

func TestUpdatePlan_InvalidID(t *testing.T) {
	_, err := client.UpdatePlan(ctx10s(t), &planpb.UpdatePlanReq{Id: "xyz", Data: newPlan("x")})
	wantCode(t, err, codes.InvalidArgument)
}

// ---------------------------------------------------------------------
// DELETE
// ---------------------------------------------------------------------

func TestDeletePlan_OK_ThenNotFound(t *testing.T) {
	created := createPlan(t, newPlan(uniqueTitle(t)))

	res, err := client.DeletePlan(ctx10s(t), &planpb.DeletePlanReq{Id: created.GetId()})
	if err != nil {
		t.Fatalf("DeletePlan: %v", err)
	}
	if !res.GetSuccess() {
		t.Errorf("success debe ser true")
	}

	_, err = client.GetPlan(ctx10s(t), &planpb.GetPlanReq{Id: created.GetId()})
	wantCode(t, err, codes.NotFound)

	// Borrar de nuevo -> NotFound.
	_, err = client.DeletePlan(ctx10s(t), &planpb.DeletePlanReq{Id: created.GetId()})
	wantCode(t, err, codes.NotFound)
}

func TestDeletePlan_NotFound(t *testing.T) {
	_, err := client.DeletePlan(ctx10s(t), &planpb.DeletePlanReq{Id: randomID()})
	wantCode(t, err, codes.NotFound)
}

func TestDeletePlan_InvalidID(t *testing.T) {
	_, err := client.DeletePlan(ctx10s(t), &planpb.DeletePlanReq{Id: "xyz"})
	wantCode(t, err, codes.InvalidArgument)
}

// ---------------------------------------------------------------------
// DELETE FEATURE (caracteristicas dentro de plan_description_items)
// ---------------------------------------------------------------------

func TestDeleteFeaturePlan_RemovesOnlyThatFeature(t *testing.T) {
	p := newPlan(uniqueTitle(t))
	p.PlanDescriptionItems = `[{"id":1,"name":"alpha"},{"id":2,"name":"beta"}]`
	created := createPlan(t, p)

	res, err := client.DeleteFeaturePlan(ctx10s(t), &planpb.DeleteFeaturePlanReq{
		Id:     1,
		PlanId: created.GetId(),
	})
	if err != nil {
		t.Fatalf("DeleteFeaturePlan: %v", err)
	}

	var items []map[string]any
	if err := json.Unmarshal([]byte(res.GetData().GetPlanDescriptionItems()), &items); err != nil {
		t.Fatalf("plan_description_items no es JSON valido: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("debe quedar 1 caracteristica, got %d", len(items))
	}
	if id, _ := items[0]["id"].(float64); id != 2 {
		t.Errorf("debe quedar la caracteristica 2, got %v", items[0]["id"])
	}
}

func TestDeleteFeaturePlan_UnknownFeatureKeepsAll(t *testing.T) {
	p := newPlan(uniqueTitle(t))
	p.PlanDescriptionItems = `[{"id":1},{"id":2}]`
	created := createPlan(t, p)

	res, err := client.DeleteFeaturePlan(ctx10s(t), &planpb.DeleteFeaturePlanReq{
		Id:     99,
		PlanId: created.GetId(),
	})
	if err != nil {
		t.Fatalf("DeleteFeaturePlan: %v", err)
	}

	var items []map[string]any
	if err := json.Unmarshal([]byte(res.GetData().GetPlanDescriptionItems()), &items); err != nil {
		t.Fatalf("JSON invalido: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("una caracteristica inexistente no debe borrar nada, quedaron %d", len(items))
	}
}

func TestDeleteFeaturePlan_PlanNotFound(t *testing.T) {
	_, err := client.DeleteFeaturePlan(ctx10s(t), &planpb.DeleteFeaturePlanReq{Id: 1, PlanId: randomID()})
	wantCode(t, err, codes.NotFound)
}

func TestDeleteFeaturePlan_InvalidItemsJSON(t *testing.T) {
	p := newPlan(uniqueTitle(t))
	p.PlanDescriptionItems = "esto-no-es-json"
	created := createPlan(t, p)

	_, err := client.DeleteFeaturePlan(ctx10s(t), &planpb.DeleteFeaturePlanReq{Id: 1, PlanId: created.GetId()})
	wantCode(t, err, codes.Internal)
}

// ---------------------------------------------------------------------
// REGLA DE NEGOCIO: un solo plan destacado
// ---------------------------------------------------------------------

func TestFeaturedRule_OnlyOneFeatured(t *testing.T) {
	skipUnless(t, "MSPLAN_IT_ALLOW_FEATURED") // modifica todos los planes destacados de la BD

	a := newPlan(uniqueTitle(t) + "-A")
	a.PlanFeatured = true
	planA := createPlan(t, a)

	b := newPlan(uniqueTitle(t) + "-B")
	b.PlanFeatured = true
	createPlan(t, b)

	got, err := client.GetPlan(ctx10s(t), &planpb.GetPlanReq{Id: planA.GetId()})
	if err != nil {
		t.Fatalf("GetPlan A: %v", err)
	}
	if got.GetData().GetPlanFeatured() {
		t.Errorf("al crear B destacado, A debe dejar de serlo")
	}
}

// ---------------------------------------------------------------------
// DEFECTOS CONOCIDOS (fallan hoy a proposito; MSPLAN_IT_DEFECTS=1)
// ---------------------------------------------------------------------

// D-04: GetPlanCompany esta declarado en plan.proto y enrutado en Kong
// (GET /api/v1/plan/company/{id}) pero el controller no lo implementa.
func TestDefect_D04_GetPlanCompanyNoImplementado(t *testing.T) {
	skipUnless(t, "MSPLAN_IT_DEFECTS")

	_, err := client.GetPlanCompany(ctx10s(t), &planpb.GetPlanCompanyReq{PlanIdCompany: randomID()})
	if status.Code(err) == codes.Unimplemented {
		t.Fatalf("D-04: GetPlanCompany responde Unimplemented")
	}
}

// D-05: UpdatePlan sobrescribe con valores vacios los campos omitidos (PUT con
// semantica de reemplazo total sin avisar), lo que borra datos del plan.
func TestDefect_D05_UpdateNoDebeBorrarCamposOmitidos(t *testing.T) {
	skipUnless(t, "MSPLAN_IT_DEFECTS")

	created := createPlan(t, newPlan(uniqueTitle(t)))

	_, err := client.UpdatePlan(ctx10s(t), &planpb.UpdatePlanReq{
		Id:   created.GetId(),
		Data: &planpb.PlanEntity{PlanTitle: created.GetPlanTitle()}, // solo titulo
	})
	if err != nil {
		t.Fatalf("UpdatePlan: %v", err)
	}

	got, err := client.GetPlan(ctx10s(t), &planpb.GetPlanReq{Id: created.GetId()})
	if err != nil {
		t.Fatalf("GetPlan: %v", err)
	}
	if got.GetData().GetPlanSubtitle() == "" || got.GetData().GetPlanPrice() == 0 {
		t.Errorf("D-05: se perdieron campos no enviados (subtitulo=%q precio=%v)",
			got.GetData().GetPlanSubtitle(), got.GetData().GetPlanPrice())
	}
}

// D-06: UpdatePlan no aplica las validaciones de Create (titulo obligatorio).
func TestDefect_D06_UpdateDebeValidarTituloObligatorio(t *testing.T) {
	skipUnless(t, "MSPLAN_IT_DEFECTS")

	created := createPlan(t, newPlan(uniqueTitle(t)))

	_, err := client.UpdatePlan(ctx10s(t), &planpb.UpdatePlanReq{
		Id:   created.GetId(),
		Data: &planpb.PlanEntity{PlanTitle: "", PlanSubtitle: "x"},
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("D-06: se esperaba InvalidArgument y se obtuvo %v", err)
	}
}

// D-07: DeleteFeaturePlan acepta id <= 0 (la validacion esta comentada) y, como
// las caracteristicas sin campo "id" se interpretan como id 0, las borra.
func TestDeleteFeaturePlan_IDZeroPreservesFeatureWithoutID(t *testing.T) {
	p := newPlan(uniqueTitle(t))
	p.PlanDescriptionItems = `[{"name":"sin-id"},{"id":0,"name":"cero"},{"id":1,"name":"ok"}]`
	created := createPlan(t, p)

	res, err := client.DeleteFeaturePlan(ctx10s(t), &planpb.DeleteFeaturePlanReq{
		Id:     0,
		PlanId: created.GetId(),
	})
	if err != nil {
		t.Fatalf("DeleteFeaturePlan con id=0: %v", err)
	}

	var items []map[string]any
	if err := json.Unmarshal([]byte(res.GetData().GetPlanDescriptionItems()), &items); err != nil {
		t.Fatalf("JSON inválido: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("deben quedar 2 características, got %d", len(items))
	}

	if name, _ := items[0]["name"].(string); name != "sin-id" {
		t.Errorf("la característica sin ID debe conservarse, got %v", items[0])
	}

	if _, exists := items[0]["id"]; exists {
		t.Errorf("la primera característica no debía tener ID, got %v", items[0]["id"])
	}

	if id, ok := items[1]["id"].(float64); !ok || id != 1 {
		t.Errorf("debe conservarse la característica con id=1, got %v", items[1]["id"])
	}
}
