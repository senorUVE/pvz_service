package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	config "github.com/senorUVE/pvz_service/configs"
	"github.com/senorUVE/pvz_service/internal/auth"
	"github.com/senorUVE/pvz_service/internal/controller"
	"github.com/senorUVE/pvz_service/internal/grpc"
	"github.com/senorUVE/pvz_service/internal/handler"
	"github.com/senorUVE/pvz_service/internal/metrics"
	"github.com/senorUVE/pvz_service/internal/repository"
	loglib "github.com/senorUVE/pvz_service/log"
	"github.com/sirupsen/logrus"

	_ "github.com/lib/pq"
)

func main() {
	logrus.SetFormatter(new(logrus.JSONFormatter))
	loglib.InitLogger(loglib.WithEnv())

	cfg, err := config.LoadConfig("./configs")
	if err != nil {
		logrus.Fatalf("Failed to load Config: %v", err)
	}

	db, err := repository.NewRepository(cfg.DBConfig)
	if err != nil {
		logrus.Fatalf("Failed to init db blyea: %v", err)
	}
	auth := auth.NewAuth(cfg.AuthConfig)
	srv := controller.NewPvzService(db, auth, cfg.ServiceConfig)

	sh := handler.NewPvzHandler(srv, auth, cfg.AppPort)

	go sh.Start()

	go func() {
		http.Handle("/metrics", metrics.PrometheusHandler())
		logrus.Info("Prometheus listening on :9000")
		if err := http.ListenAndServe(":9000", nil); err != nil {
			logrus.Fatalf("promethes server error: %v", err)
		}
	}()

	go func() {
		if err := grpc.StartGrpcServer(db); err != nil {
			logrus.Fatalf("failed to start grpc server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := db.Close(); err != nil {
		logrus.Fatal(err)
	}
	if err := sh.Close(ctx); err != nil {
		logrus.Fatal(err)
	}

	select {
	case <-ctx.Done():
		logrus.Println("timeout of 3 seconds")
	}
	logrus.Println("Server exiting")
}
