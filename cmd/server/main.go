package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/rusik69/serverscheduler/internal/config"
	"github.com/rusik69/serverscheduler/internal/database"
	"github.com/rusik69/serverscheduler/internal/logger"
	"github.com/rusik69/serverscheduler/internal/server"
	"github.com/rusik69/serverscheduler/internal/services"
)

func main() {
	_ = godotenv.Load()
	cfg := config.LoadConfig()
	logger.Init(cfg.LogLevel)

	db, err := database.InitDB(cfg.DBPath)
	if err != nil {
		slog.Error("database init failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	userSvc := services.NewUserService(db)
	serverSvc := services.NewServerService(db)
	resSvc := services.NewReservationService(db)
	sshSvc := services.NewSSHService()
	slackSvc := services.NewSlackService(cfg.SlackWebhookURL)

	srv := server.NewServer(cfg, userSvc, serverSvc, resSvc, sshSvc, slackSvc)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		cancel()
	}()

	done := make(chan error, 1)
	go func() { done <- srv.Start(ctx) }()
	select {
	case err := <-done:
		if err != nil {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	case <-ctx.Done():
		slog.Info("shutting down")
	}
}
