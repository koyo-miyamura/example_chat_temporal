package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

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

	updateInput := UserReplyInput{
		Text:   r.URL.Query().Get("data"),
		UserID: userID,
		ChatID: chatID,
	}

	result, err := UpdateWorkflowWithUserMessage(r.Context(), updateInput)
	if err != nil {
		if errors.Is(err, ErrChatAlreadyCompleted) {
			http.Error(w, "Chat already completed", http.StatusConflict)
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
