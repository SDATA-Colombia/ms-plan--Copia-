package mongodb

import (
	"context"
	"sync"
	"time"

	"github.com/SDATA-Colombia/ms-plan/internal/config"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Mongo struct {
	client   *mongo.Client
	database *mongo.Database
}

var (
	instance *Mongo
	once     sync.Once
)

func Init(cfg *config.Config) error {

	var err error

	once.Do(func() {

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client, e := mongo.Connect(
			options.Client().ApplyURI(cfg.MongoURI),
		)

		if e != nil {
			err = e
			return
		}

		if e = client.Ping(ctx, nil); e != nil {
			err = e
			return
		}

		instance = &Mongo{
			client:   client,
			database: client.Database(cfg.MongoDatabase),
		}
	})

	return err
}

func Close() {

	if instance == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = instance.client.Disconnect(ctx)
}

func Client() *mongo.Client {
	if instance == nil {
		return nil
	}

	return instance.client
}

func Database() *mongo.Database {
	if instance == nil {
		return nil
	}

	return instance.database
}

func Collection(name string) *mongo.Collection {
	if instance == nil || instance.database == nil {
		return nil
	}

	return instance.database.Collection(name)
}
