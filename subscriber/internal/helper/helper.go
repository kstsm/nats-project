package helper

import (
	"encoding/json"
	"github.com/gookit/slog"
	"net/http"
)

func ResponseJson(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Ошибка при кодировании JSON", http.StatusInternalServerError)
		slog.Error("Ошибка JSON кодирования", err)
	}
}
