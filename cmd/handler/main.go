package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/koyo-miyamura/example_chat_temporal/internal/chat"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	go func() {
		slog.Info("API server started...")
		if err := http.ListenAndServe(":8080", newServer()); err != nil {
			slog.Error(fmt.Sprintf("Unable to start server: %v", err))
		}
	}()

	<-ctx.Done()
	slog.Info("shutdown")
}

func newServer() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/chat", handleChat)

	return mux
}

// POST /chat?chat_id=xxx&user_id=yyy&data=zzz
func handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	chatID := r.URL.Query().Get("chat_id")
	if chatID == "" {
		http.Error(w, "chat_id is required", http.StatusBadRequest)
		return
	}
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	updateInput := chat.UserReplyInput{
		Text:   r.URL.Query().Get("data"),
		UserID: userID,
		ChatID: chatID,
	}

	result, err := chat.UpdateWorkflowWithUserMessage(r.Context(), updateInput)
	if err != nil {
		if errors.Is(err, chat.ErrChatAlreadyCompleted) {
			http.Error(w, "Chat already completed", http.StatusConflict)
			return
		}
		if errors.Is(err, chat.ErrChatProcessing) {
			http.Error(w, "Chat is currently processing, please wait", http.StatusConflict)
			return
		}
		slog.Error(fmt.Sprintf("[handleChat] UpdateWorkflowWithUserMessage error: %v", err))
		http.Error(w, fmt.Sprintf("Workflow update error: %v", err), http.StatusInternalServerError)
		return
	}

	slog.Info(fmt.Sprintf("[handleChat] WorkflowID: %s message accepted", chatID))
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"chat_id": chatID, "status": result}); err != nil {
		slog.Error(fmt.Sprintf("[handleChat] JSON encode error: %v", err))
		http.Error(w, fmt.Sprintf("JSON encode error: %v", err), http.StatusInternalServerError)
		return
	}
}
