package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.temporal.io/sdk/worker"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	temporalClient, err := temporalClient()
	if err != nil {
		slog.Error(fmt.Sprintf("Unable to create client: %v", err))
		os.Exit(1)
	}
	defer temporalClient.Close()

	go startWorker()

	go startAPIServer()

	<-ctx.Done()
	slog.Info("shutdown")
}

func startWorker() {
	temporalClient, err := temporalClient()
	if err != nil {
		slog.Error(fmt.Sprintf("Unable to create client: %v", err))
		os.Exit(1)
	}
	defer temporalClient.Close()

	w := worker.New(temporalClient, ChatTaskQueue, worker.Options{})
	w.RegisterWorkflow(ChatWorkflow)
	w.RegisterActivity(ChatStep1Activity)
	w.RegisterActivity(ChatStep2Activity)

	slog.Info("Workers started...")

	if err := w.Run(worker.InterruptCh()); err != nil {
		slog.Error(fmt.Sprintf("Unable to start worker: %v", err))
	}
}

func startAPIServer() {
	server := newServer()

	slog.Info("API server started...")
	if err := http.ListenAndServe(":8080", server); err != nil {
		slog.Error(fmt.Sprintf("Unable to start gateway: %v", err))
		return
	}
}
