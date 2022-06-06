package main

import (
	"ilmavridis/url-shortener/config"
	"ilmavridis/url-shortener/logger"
	"ilmavridis/url-shortener/redisStorage"
	"ilmavridis/url-shortener/routes"

	"log"
	"os"
	"os/signal"

	"go.uber.org/zap"
)

func main() {

	err := logger.New()
	if err != nil {
		log.Printf("Can't initialize zap logger: %v", err)
	}
	loggerZap := logger.Get()
	defer loggerZap.Sync()

	err = config.Read()
	conf := config.Get()
	if err != nil {
		logger.Fatal("Error reading configuration: ", err)
	}
	logger.Info("Server Configuration",
		zap.String("address", conf.Server.Address),
		zap.Duration("write timeout", conf.Server.TimeoutWrite),
		zap.Duration("read timeout", conf.Server.TimeoutRead),
		zap.Duration("idle timeout", conf.Server.TimeoutIdle),
	)

	err = redisStorage.CreateClient()
	if err != nil {
		logger.Fatal("Could not connect to redis: ", err)
	}
	redisClient := redisStorage.Get()
	defer redisClient.Close()
	logger.Info("Connected to redis", zap.String("address", conf.Redis.Address))

	srv := routes.New()
	errs := routes.Run(srv)
	logger.Info("Server start running, listening at ", zap.String("address", srv.Addr))

	// Graceful shutdown when recieving SIGINT / Ctrl+C
	// SIGKILL, SIGQUIT or SIGTERM will not be caught
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)
	select {
	case err := <-errs:
		logger.Error("Error starting server: ", err)
	case sig := <-stopChan:
		logger.Info("Signal received, shutting down server...", zap.String("signal", sig.String()))
		if err := routes.SetupGracefulShutdown(srv); err != nil {
			logger.Error("Server shutdown error: ", err)
		}
		logger.Info("Server gracefully stopped!")
		os.Exit(0)
	}

}
