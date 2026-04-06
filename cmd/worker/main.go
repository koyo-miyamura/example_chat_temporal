package main

import (
	"fmt"
	"log/slog"
	"os"

	"go.temporal.io/sdk/worker"

	"github.com/koyo-miyamura/example_chat_temporal/internal/chat"
)

func main() {
	temporalClient, err := chat.NewClient()
	if err != nil {
		slog.Error(fmt.Sprintf("Unable to create client: %v", err))
		os.Exit(1)
	}
	defer temporalClient.Close()

	w := worker.New(temporalClient, chat.ChatTaskQueue, worker.Options{})
	w.RegisterWorkflow(chat.ChatWorkflow)
	w.RegisterActivity(chat.ChatStep1Activity)
	w.RegisterActivity(chat.ChatStep2Activity)

	slog.Info("Workers started...")

	if err := w.Run(worker.InterruptCh()); err != nil {
		slog.Error(fmt.Sprintf("Unable to start worker: %v", err))
	}
}
