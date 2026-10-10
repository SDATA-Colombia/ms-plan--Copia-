package controller

// Pruebas del controller gRPC SIN infraestructura (sin MongoDB, Dragonfly ni Pulsar).
//
// Se levanta un servidor gRPC real en memoria (bufconn) con PlanController y
// service = nil. Solo se prueban los caminos que terminan ANTES de llamar al
// service: validacion de IDs y conversion proto -> entity. Asi se verifica
// tambien que los codigos gRPC (codes.InvalidArgument, etc.) viajan por el cable.

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	planpb "github.com/SDATA-Colombia/shared-go-proto/plan/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func newTestClient(t *testing.T) planpb.PlanServiceClient {
	t.Helper()

	lis := bufconn.Listen(1024 * 1024)

	srv := grpc.NewServer()
	planpb.RegisterPlanServiceServer(srv, NewPlanController(nil))

	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("no se pudo crear el cliente gRPC: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	return planpb.NewPlanServiceClient(conn)
}

func ctx5s(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	return ctx
}

const validHexID = "64b7f0a1c2d3e4f5a6b7c8d9"

func TestPlanController_Validation(t *testing.T) {
	cl := newTestClient(t)

	cases := []struct {
		name     string
		call     func(ctx context.Context) error
		wantCode codes.Code
		wantMsg  string
	}{
		// ---------------- GetPlan ----------------
		{
			name: "GetPlan id vacio",
			call: func(ctx context.Context) error {
				_, err := cl.GetPlan(ctx, &planpb.GetPlanReq{Id: ""})
				return err
			},
			wantCode: codes.InvalidArgument, wantMsg: "ID is required",
		},
		{
			name: "GetPlan id solo espacios",
			call: func(ctx context.Context) error {
				_, err := cl.GetPlan(ctx, &planpb.GetPlanReq{Id: "   "})
				return err
			},
			wantCode: codes.InvalidArgument, wantMsg: "ID is required",
		},
		{
			name: "GetPlan id con formato invalido",
			call: func(ctx context.Context) error {
				_, err := cl.GetPlan(ctx, &planpb.GetPlanReq{Id: "no-es-objectid"})
				return err
			},
			wantCode: codes.InvalidArgument, wantMsg: "Invalid ID format",
		},
		{
			name: "GetPlan id de 23 caracteres",
			call: func(ctx context.Context) error {
				_, err := cl.GetPlan(ctx, &planpb.GetPlanReq{Id: validHexID[:23]})
				return err
			},
			wantCode: codes.InvalidArgument, wantMsg: "Invalid ID format",
		},

		// ---------------- DeletePlan ----------------
		{
			name: "DeletePlan id vacio",
			call: func(ctx context.Context) error {
				_, err := cl.DeletePlan(ctx, &planpb.DeletePlanReq{Id: ""})
				return err
			},
			wantCode: codes.InvalidArgument, wantMsg: "ID is required",
		},
		{
			name: "DeletePlan id invalido",
			call: func(ctx context.Context) error {
				_, err := cl.DeletePlan(ctx, &planpb.DeletePlanReq{Id: "xyz"})
				return err
			},
			wantCode: codes.InvalidArgument, wantMsg: "Invalid ID format",
		},

		// ---------------- UpdatePlan ----------------
		{
			name: "UpdatePlan id vacio",
			call: func(ctx context.Context) error {
				_, err := cl.UpdatePlan(ctx, &planpb.UpdatePlanReq{Id: ""})
				return err
			},
			wantCode: codes.InvalidArgument, wantMsg: "ID is required",
		},
		{
			name: "UpdatePlan id invalido",
			call: func(ctx context.Context) error {
				_, err := cl.UpdatePlan(ctx, &planpb.UpdatePlanReq{Id: "xyz"})
				return err
			},
			wantCode: codes.InvalidArgument, wantMsg: "Invalid ID format",
		},
		{
			name: "UpdatePlan data.id invalido",
			call: func(ctx context.Context) error {
				_, err := cl.UpdatePlan(ctx, &planpb.UpdatePlanReq{
					Id:   validHexID,
					Data: &planpb.PlanEntity{Id: "malo"},
				})
				return err
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "UpdatePlan data.plan_id_company invalido",
			call: func(ctx context.Context) error {
				_, err := cl.UpdatePlan(ctx, &planpb.UpdatePlanReq{
					Id:   validHexID,
					Data: &planpb.PlanEntity{PlanIdCompany: "malo"},
				})
				return err
			},
			wantCode: codes.InvalidArgument,
		},

		// ---------------- CreatePlan ----------------
		{
			name: "CreatePlan data.id invalido",
			call: func(ctx context.Context) error {
				_, err := cl.CreatePlan(ctx, &planpb.CreatePlanReq{
					Data: &planpb.PlanEntity{Id: "malo", PlanTitle: "T", PlanSubtitle: "S"},
				})
				return err
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "CreatePlan data.plan_id_company invalido",
			call: func(ctx context.Context) error {
				_, err := cl.CreatePlan(ctx, &planpb.CreatePlanReq{
					Data: &planpb.PlanEntity{PlanIdCompany: "malo", PlanTitle: "T", PlanSubtitle: "S"},
				})
				return err
			},
			wantCode: codes.InvalidArgument,
		},

		// ---------------- DeleteFeaturePlan ----------------
		{
			name: "DeleteFeaturePlan plan_id vacio",
			call: func(ctx context.Context) error {
				_, err := cl.DeleteFeaturePlan(ctx, &planpb.DeleteFeaturePlanReq{Id: 1, PlanId: ""})
				return err
			},
			wantCode: codes.InvalidArgument, wantMsg: "Plan ID is required",
		},
		{
			name: "DeleteFeaturePlan plan_id solo espacios",
			call: func(ctx context.Context) error {
				_, err := cl.DeleteFeaturePlan(ctx, &planpb.DeleteFeaturePlanReq{Id: 1, PlanId: "  "})
				return err
			},
			wantCode: codes.InvalidArgument, wantMsg: "Plan ID is required",
		},
		{
			name: "DeleteFeaturePlan plan_id invalido",
			call: func(ctx context.Context) error {
				_, err := cl.DeleteFeaturePlan(ctx, &planpb.DeleteFeaturePlanReq{Id: 1, PlanId: "xyz"})
				return err
			},
			wantCode: codes.InvalidArgument, wantMsg: "Invalid ID format",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.call(ctx5s(t))
			if err == nil {
				t.Fatalf("se esperaba error con codigo %s y no hubo error", c.wantCode)
			}

			st, ok := status.FromError(err)
			if !ok {
				t.Fatalf("el error no es un status gRPC: %v", err)
			}
			if st.Code() != c.wantCode {
				t.Errorf("codigo: got %s want %s (msg=%q)", st.Code(), c.wantCode, st.Message())
			}
			if c.wantMsg != "" && !strings.Contains(st.Message(), c.wantMsg) {
				t.Errorf("mensaje: got %q, debe contener %q", st.Message(), c.wantMsg)
			}
		})
	}
}

// El constructor debe conservar el service recibido.
func TestNewPlanController_NotNil(t *testing.T) {
	c := NewPlanController(nil)
	if c == nil {
		t.Fatal("NewPlanController no debe devolver nil")
	}
}
