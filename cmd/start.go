package cmd

// Start: Punto de entrada del microservicio.
// Inicializa configuraciones, conexiones (MongoDB, Dragonfly, Pulsar),
// registra servicios gRPC y maneja el ciclo de vida (apagado limpio).

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/SDATA-Colombia/ms-plan/indexes"
	"github.com/SDATA-Colombia/ms-plan/internal/config"
	"github.com/SDATA-Colombia/ms-plan/internal/consumer"
	"github.com/SDATA-Colombia/ms-plan/internal/controller"
	"github.com/SDATA-Colombia/ms-plan/internal/model/entity"
	"github.com/SDATA-Colombia/ms-plan/internal/repository"
	"github.com/SDATA-Colombia/ms-plan/internal/service"

	"github.com/SDATA-Colombia/ms-plan/pkg/dragonfly"
	"github.com/SDATA-Colombia/ms-plan/pkg/logger"
	"github.com/SDATA-Colombia/ms-plan/pkg/mongodb"
	"github.com/SDATA-Colombia/ms-plan/pkg/pulsar"

	grpcserver "github.com/SDATA-Colombia/ms-plan/pkg/grpc"
	gogrpc "google.golang.org/grpc"
)

func Start() {
	cfg := config.Load()

	logger.Init(cfg.Environment)

	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("      SDATA Microservice ms-plan")
	fmt.Println("========================================")
	fmt.Println()

	logger.Info("Configuration loaded")
	logger.Info("Environment: " + cfg.Environment)

	//---------------------------------------------------
	// MongoDB
	//---------------------------------------------------

	if err := mongodb.Init(cfg); err != nil {
		logger.Fatal(err.Error())
	}

	logger.Info("MongoDB connected")

	//---------------------------------------------------
	// MongoDB Indexes
	//---------------------------------------------------

	if err := indexes.Init(context.Background()); err != nil {
		logger.Fatal(err.Error())
	}

	logger.Info("MongoDB indexes created")

	//---------------------------------------------------
	// Dragonfly
	//---------------------------------------------------

	if err := dragonfly.Init(cfg); err != nil {
		logger.Fatal(err.Error())
	}

	logger.Info("Dragonfly connected")

	//---------------------------------------------------
	// Apache Pulsar
	//---------------------------------------------------

	if err := pulsar.Init(cfg); err != nil {
		logger.Fatal(err.Error())
	}

	logger.Info("Apache Pulsar connected")

	//---------------------------------------------------
	// Registrar Consumers
	//---------------------------------------------------

	if err := consumer.Register(); err != nil {
		logger.Fatal(err.Error())
	}

	//---------------------------------------------------
	// Inicialización de Repositorios y Servicios
	//---------------------------------------------------

	baseRepo := repository.NewBaseRepository[entity.Plan]("plans")
	baseService := service.NewBaseService(baseRepo)

	//---------------------------------------------------
	// gRPC
	//---------------------------------------------------

	server := grpcserver.New(cfg.GRPCPort)

	//---------------------------------------------------
	// Registrar servicios gRPC
	//---------------------------------------------------

	server.Register(func(s *gogrpc.Server) {
		controller.Register(
			s,
			baseService,
		)
	})

	go func() {
		if err := server.Start(); err != nil {
			logger.Fatal(err.Error())
		}
	}()

	logger.Info("gRPC Server started")

	//---------------------------------------------------
	// Esperar señal del sistema
	//---------------------------------------------------

	quit := make(chan os.Signal, 1)

	signal.Notify(
		quit,
		os.Interrupt,
		syscall.SIGTERM,
	)

	<-quit

	//---------------------------------------------------
	// Apagado limpio del microservicio
	//---------------------------------------------------

	logger.Info("Stopping microservice...")
	server.Stop()
	pulsar.Close()
	mongodb.Close()
	dragonfly.Close()
	logger.Info("Microservice stopped")
}
