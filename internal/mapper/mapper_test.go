package mapper

import (
	"testing"
	"time"

	"github.com/SDATA-Colombia/ms-plan/internal/model/entity"
	planpb "github.com/SDATA-Colombia/shared-go-proto/plan/v1"

	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/protobuf/types/known/timestamppb"
	"reflect"
)

var baseTime = time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)

func fullProto(id, company bson.ObjectID) *planpb.PlanEntity {
	return &planpb.PlanEntity{
		Id:                     id.Hex(),
		PlanIdCompany:          company.Hex(),
		PlanAccentColor:        "#112233",
		PlanAccentTextColor:    "#ffffff",
		PlanTitle:              "Pro",
		PlanSubtitle:           "Plan profesional",
		PlanDescriptionTinymce: "<p>descripcion</p>",
		PlanType:               "monthly",
		PlanCurrencyPrice:      "COP",
		PlanMenu:               "main",
		PlanDescriptionItems:   `[{"id":1,"name":"a"}]`,
		PlanDays:               30,
		PlanOffices:            2,
		PlanUsers:              10,
		PlanTotalSales:         7,
		PlanPrice:              99900.5,
		PlanFeatured:           true,
		PlanCustom:             true,
		ActivePlan:             true,
		CreatedAt:              timestamppb.New(baseTime),
		UpdatedAt:              timestamppb.New(baseTime.Add(time.Hour)),
	}
}

// ---------------------------------------------------------------------
// Proto -> Entity
// ---------------------------------------------------------------------

func TestProtoToEntity_MapsAllFields(t *testing.T) {
	id := bson.NewObjectID()
	company := bson.NewObjectID()

	got, err := ProtoToEntity(fullProto(id, company))
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if got.ID != id {
		t.Errorf("ID: got %s want %s", got.ID.Hex(), id.Hex())
	}
	if got.PlanIDCompany != company {
		t.Errorf("PlanIDCompany: got %s want %s", got.PlanIDCompany.Hex(), company.Hex())
	}

	checks := []struct {
		name      string
		got, want any
	}{
		{"PlanAccentColor", got.PlanAccentColor, "#112233"},
		{"PlanAccentTextColor", got.PlanAccentTextColor, "#ffffff"},
		{"PlanTitle", got.PlanTitle, "Pro"},
		{"PlanSubtitle", got.PlanSubtitle, "Plan profesional"},
		{"PlanDescriptionTinyMCE", got.PlanDescriptionTinyMCE, "<p>descripcion</p>"},
		{"PlanType", got.PlanType, "monthly"},
		{"PlanCurrencyPrice", got.PlanCurrencyPrice, "COP"},
		{"PlanMenu", got.PlanMenu, "main"},
		{"PlanDescriptionItems", got.PlanDescriptionItems, `[{"id":1,"name":"a"}]`},
		{"PlanDays", got.PlanDays, int32(30)},
		{"PlanOffices", got.PlanOffices, int32(2)},
		{"PlanUsers", got.PlanUsers, int32(10)},
		{"PlanTotalSales", got.PlanTotalSales, int32(7)},
		{"PlanPrice", got.PlanPrice, 99900.5},
		{"PlanFeatured", got.PlanFeatured, true},
		{"PlanCustom", got.PlanCustom, true},
		{"ActivePlan", got.ActivePlan, true},
	}

	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s: got %v want %v", c.name, c.got, c.want)
		}
	}

	if !got.CreatedAt.Equal(baseTime) {
		t.Errorf("CreatedAt: got %v want %v", got.CreatedAt, baseTime)
	}
	if !got.UpdatedAt.Equal(baseTime.Add(time.Hour)) {
		t.Errorf("UpdatedAt: got %v", got.UpdatedAt)
	}
	if got.DeletedAt != nil {
		t.Errorf("DeletedAt debe ser nil cuando el proto no lo trae")
	}
}

func TestProtoToEntity_EmptyIDsStayZero(t *testing.T) {
	got, err := ProtoToEntity(&planpb.PlanEntity{PlanTitle: "Solo titulo"})
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if !got.ID.IsZero() {
		t.Errorf("ID debe quedar en cero")
	}
	if !got.PlanIDCompany.IsZero() {
		t.Errorf("PlanIDCompany debe quedar en cero")
	}
}

func TestProtoToEntity_NilProto(t *testing.T) {
	// Los getters de protobuf son seguros ante nil.
	got, err := ProtoToEntity(nil)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if got == nil || got.PlanTitle != "" {
		t.Errorf("se esperaba una entidad vacia, got %+v", got)
	}
}

func TestProtoToEntity_InvalidIDs(t *testing.T) {
	cases := []struct {
		name  string
		input *planpb.PlanEntity
	}{
		{"id no hex", &planpb.PlanEntity{Id: "no-es-un-objectid"}},
		{"id corto", &planpb.PlanEntity{Id: "abc123"}},
		{"id con 24 chars no hex", &planpb.PlanEntity{Id: "zzzzzzzzzzzzzzzzzzzzzzzz"}},
		{"company no hex", &planpb.PlanEntity{PlanIdCompany: "company-invalida"}},
		{"company corto", &planpb.PlanEntity{PlanIdCompany: "123"}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ProtoToEntity(c.input)
			if err == nil {
				t.Fatalf("se esperaba error y se obtuvo %+v", got)
			}
			if got != nil {
				t.Errorf("con error la entidad debe ser nil")
			}
		})
	}
}

func TestProtoToEntity_DeletedAt(t *testing.T) {
	in := &planpb.PlanEntity{DeletedAt: timestamppb.New(baseTime)}

	got, err := ProtoToEntity(in)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if got.DeletedAt == nil || !got.DeletedAt.Equal(baseTime) {
		t.Errorf("DeletedAt: got %v want %v", got.DeletedAt, baseTime)
	}
}

// ---------------------------------------------------------------------
// Entity -> Proto
// ---------------------------------------------------------------------

func TestEntityToProto_MapsAllFields(t *testing.T) {
	deleted := baseTime.Add(2 * time.Hour)
	in := &entity.Plan{
		ID:                     bson.NewObjectID(),
		PlanIDCompany:          bson.NewObjectID(),
		PlanAccentColor:        "#000",
		PlanAccentTextColor:    "#fff",
		PlanTitle:              "Basic",
		PlanSubtitle:           "Basico",
		PlanDescriptionTinyMCE: "<b>x</b>",
		PlanType:               "annual",
		PlanCurrencyPrice:      "USD",
		PlanMenu:               "side",
		PlanDescriptionItems:   "[]",
		PlanDays:               365,
		PlanOffices:            1,
		PlanUsers:              3,
		PlanTotalSales:         5,
		PlanPrice:              19.99,
		PlanFeatured:           true,
		PlanCustom:             false,
		ActivePlan:             true,
		CreatedAt:              baseTime,
		UpdatedAt:              baseTime.Add(time.Minute),
		DeletedAt:              &deleted,
	}

	got := EntityToProto(in)

	if got.GetId() != in.ID.Hex() {
		t.Errorf("Id: got %s want %s", got.GetId(), in.ID.Hex())
	}
	if got.GetPlanIdCompany() != in.PlanIDCompany.Hex() {
		t.Errorf("PlanIdCompany: got %s", got.GetPlanIdCompany())
	}
	if got.GetPlanTitle() != "Basic" || got.GetPlanSubtitle() != "Basico" {
		t.Errorf("titulo/subtitulo incorrectos: %q / %q", got.GetPlanTitle(), got.GetPlanSubtitle())
	}
	if got.GetPlanDays() != 365 || got.GetPlanOffices() != 1 || got.GetPlanUsers() != 3 || got.GetPlanTotalSales() != 5 {
		t.Errorf("campos enteros incorrectos: %+v", got)
	}
	if got.GetPlanPrice() != 19.99 {
		t.Errorf("PlanPrice: got %v", got.GetPlanPrice())
	}
	if !got.GetPlanFeatured() || got.GetPlanCustom() || !got.GetActivePlan() {
		t.Errorf("banderas incorrectas: featured=%v custom=%v active=%v",
			got.GetPlanFeatured(), got.GetPlanCustom(), got.GetActivePlan())
	}
	if !got.GetCreatedAt().AsTime().Equal(baseTime) {
		t.Errorf("CreatedAt: got %v", got.GetCreatedAt().AsTime())
	}
	if got.GetDeletedAt() == nil || !got.GetDeletedAt().AsTime().Equal(deleted) {
		t.Errorf("DeletedAt: got %v want %v", got.GetDeletedAt(), deleted)
	}
}

func TestEntityToProto_DeletedAtNil(t *testing.T) {
	got := EntityToProto(&entity.Plan{ID: bson.NewObjectID(), CreatedAt: baseTime, UpdatedAt: baseTime})
	if got.GetDeletedAt() != nil {
		t.Errorf("DeletedAt debe ser nil cuando la entidad no esta borrada")
	}
}

// ---------------------------------------------------------------------
// Round trip y listas
// ---------------------------------------------------------------------

func TestRoundTrip_ProtoEntityProto(t *testing.T) {
	in := fullProto(bson.NewObjectID(), bson.NewObjectID())

	ent, err := ProtoToEntity(in)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	out := EntityToProto(ent)

	if !reflect.DeepEqual(in, out) {
		t.Errorf("el round trip no conserva los datos\n in: %v\nout: %v", in, out)
	}
}

func TestEntitiesToProto(t *testing.T) {
	t.Run("nil devuelve lista vacia no nula", func(t *testing.T) {
		got := EntitiesToProto(nil)
		if got == nil {
			t.Fatalf("se esperaba slice no nil para serializar [] y no null")
		}
		if len(got) != 0 {
			t.Errorf("len: got %d want 0", len(got))
		}
	})

	t.Run("conserva cantidad y orden", func(t *testing.T) {
		a := &entity.Plan{ID: bson.NewObjectID(), PlanTitle: "A", CreatedAt: baseTime, UpdatedAt: baseTime}
		b := &entity.Plan{ID: bson.NewObjectID(), PlanTitle: "B", CreatedAt: baseTime, UpdatedAt: baseTime}
		c := &entity.Plan{ID: bson.NewObjectID(), PlanTitle: "C", CreatedAt: baseTime, UpdatedAt: baseTime}

		got := EntitiesToProto([]*entity.Plan{a, b, c})

		if len(got) != 3 {
			t.Fatalf("len: got %d want 3", len(got))
		}
		for i, want := range []string{"A", "B", "C"} {
			if got[i].GetPlanTitle() != want {
				t.Errorf("posicion %d: got %s want %s", i, got[i].GetPlanTitle(), want)
			}
		}
	})
}
