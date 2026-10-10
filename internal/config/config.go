package config

import (
	"os"
	"strconv"
)

type Config struct {
	GRPCPort      int
	Environment   string
	MongoURI      string
	MongoDatabase string
	DragonflyAddr string
	DragonflyPass string
	PulsarURL     string
	PulsarToken   string
}

func Load() *Config {
	grpcPort, _ := strconv.Atoi(os.Getenv("GRPC_PORT"))

	return &Config{
		GRPCPort:      grpcPort,
		Environment:   os.Getenv("APP_ENV"),
		MongoURI:      os.Getenv("MONGO_URI"),
		MongoDatabase: os.Getenv("MONGO_DATABASE"),
		DragonflyAddr: os.Getenv("DRAGONFLY_ADDR"),
		DragonflyPass: os.Getenv("DRAGONFLY_PASS"),
		PulsarURL:     os.Getenv("PULSAR_URL"),
		PulsarToken:   os.Getenv("PULSAR_TOKEN"),
	}
}

func (c *Config) IsDevelopment() bool {
	return c != nil &&
		(c.Environment == "development" || c.Environment == "dev")
}
