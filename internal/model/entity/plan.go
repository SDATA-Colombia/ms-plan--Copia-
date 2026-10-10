package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Plan representa la colección 'plans' en MongoDB.
// Usamos 'omitempty' para emular el 'required: false' de Mongoose.
type Plan struct {
	ID                     bson.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	PlanIDCompany          bson.ObjectID `bson:"plan_id_company,omitempty" json:"plan_id_company,omitempty"`
	PlanAccentColor        string        `bson:"plan_accent_color,omitempty" json:"plan_accent_color,omitempty"`
	PlanAccentTextColor    string        `bson:"plan_accent_text_color,omitempty" json:"plan_accent_text_color,omitempty"`
	PlanTitle              string        `bson:"plan_title,omitempty" json:"plan_title,omitempty"`
	PlanSubtitle           string        `bson:"plan_subtitle,omitempty" json:"plan_subtitle,omitempty"`
	PlanDescriptionTinyMCE string        `bson:"plan_description_tinymce,omitempty" json:"plan_description_tinymce,omitempty"`
	PlanType               string        `bson:"plan_type,omitempty" json:"plan_type,omitempty"`
	PlanCurrencyPrice      string        `bson:"plan_currency_price,omitempty" json:"plan_currency_price,omitempty"`
	PlanMenu               string        `bson:"plan_menu,omitempty" json:"plan_menu,omitempty"`
	PlanDescriptionItems   string        `bson:"plan_description_items,omitempty" json:"plan_description_items,omitempty"`

	PlanDays       int32   `bson:"plan_days,omitempty" json:"plan_days,omitempty"`
	PlanOffices    int32   `bson:"plan_offices,omitempty" json:"plan_offices,omitempty"`
	PlanUsers      int32   `bson:"plan_users,omitempty" json:"plan_users,omitempty"`
	PlanTotalSales int32   `bson:"plan_total_sales,omitempty" json:"plan_total_sales,omitempty"`
	PlanPrice      float64 `bson:"plan_price,omitempty" json:"plan_price,omitempty"`

	PlanFeatured bool `bson:"plan_featured,omitempty" json:"plan_featured,omitempty"`
	PlanCustom   bool `bson:"plan_custom,omitempty" json:"plan_custom,omitempty"`
	ActivePlan   bool `bson:"active_plan,omitempty" json:"active_plan,omitempty"`

	CreatedAt time.Time  `bson:"createdAt,omitempty" json:"createdAt,omitempty"`
	UpdatedAt time.Time  `bson:"updatedAt,omitempty" json:"updatedAt,omitempty"`
	DeletedAt *time.Time `bson:"deletedAt,omitempty" json:"deletedAt,omitempty"`
}
