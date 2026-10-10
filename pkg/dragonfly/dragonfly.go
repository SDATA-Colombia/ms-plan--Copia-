package dragonfly

import (
	"context"
	"sync"
	"time"

	"github.com/SDATA-Colombia/ms-plan/internal/config"
	"github.com/redis/go-redis/v9"
)

type Dragonfly struct {
	client *redis.Client
}

var (
	instance *Dragonfly
	once     sync.Once
)

func Init(cfg *config.Config) error {

	var err error

	once.Do(func() {

		client := redis.NewClient(&redis.Options{
			Addr:     cfg.DragonflyAddr,
			Password: cfg.DragonflyPass,
			DB:       0,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if e := client.Ping(ctx).Err(); e != nil {
			err = e
			return
		}

		instance = &Dragonfly{
			client: client,
		}
	})

	return err
}

func Close() {

	if instance == nil {
		return
	}

	_ = instance.client.Close()
}

func Client() *redis.Client {
	return instance.client
}
