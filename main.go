package main

import (
	"github.com/Netflix/go-env"
	"github.com/tonx22/playlist/pkg/service"
	"github.com/tonx22/playlist/pkg/transport"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	_ "github.com/lib/pq"
)

type environment struct {
	PgsqlURI string `env:"POSTGRES_URI"`
	GRPCPort int    `env:"GRPC_PORT"`
}

func main() {

	var e environment
	_, err := env.UnmarshalFromEnviron(&e)
	if err != nil {
		log.Fatalf("Can't get environment variables: %v", err)
	}

	svc, err := service.NewPlaylistService(e.PgsqlURI)
	if err != nil {
		log.Fatalf("Can't create PlaylistService: %v", err)
	}

	// Running migrations
	driver, err := postgres.WithInstance(svc.DB, &postgres.Config{})
	if err != nil {
		log.Fatalf("Can't get postgres driver: %v", err)
	}
	m, err := migrate.NewWithDatabaseInstance("file://./migrations", "postgres", driver)
	if err != nil {
		log.Fatalf("Can't get migration object: %v", err)
	}
	m.Up()

	err = transport.StartNewGRPCServer(svc, e.GRPCPort)
	if err != nil {
		log.Fatalf("Failed to start GRPC server: %v", err)
	}

	var sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, os.Interrupt)
	<-sigChan
	service.Shutdown(svc)
}
