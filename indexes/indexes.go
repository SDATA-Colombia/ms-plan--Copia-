package indexes

import (
	"context"
	"github.com/SDATA-Colombia/ms-plan/internal/model/entity"
	"github.com/SDATA-Colombia/ms-plan/internal/repository"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Init inicializa los índices de la colección "plans".
func Init(ctx context.Context) error {
	repo := repository.NewBaseRepository[entity.Plan]("plans")
	return repo.EnsureIndexes(ctx, PlanIndexes())
}

// PlanIndexes retorna todos los índices de la colección "plans".
func PlanIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "plan_title", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	}
}
