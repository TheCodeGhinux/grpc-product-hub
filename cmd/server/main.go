package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"grpc-product/internal/server"
)

func gracefulShutdown(apiServer *http.Server, done chan bool) {

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()


	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")
	stop()



	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	log.Println("Server exiting")


	done <- true
}

func main() {

	server.StartGRPCServer()


	done := make(chan bool, 1)




	<-done
	log.Println("Graceful shutdown complete.")
}
