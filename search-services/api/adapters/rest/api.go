package rest

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"yadro.com/course/api/core"
)

type pingResponse struct {
	Replies map[string]string `json:"replies"`
}

type wordsResponse struct {
	Words []string `json:"words"`
	Total int      `json:"total"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func NewPingHandler(log *slog.Logger, timeout time.Duration, pingers map[string]core.Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		replies := make(map[string]string, len(pingers))

		for name, pinger := range pingers {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			err := pinger.Ping(ctx)
			cancel()

			if err != nil {
				log.Warn("ping failed", "service", name, "error", err)
				replies[name] = "unavailable"
				continue
			}

			replies[name] = "ok"
		}

		writeJSON(w, http.StatusOK, pingResponse{Replies: replies})
	}
}

func NewWordsHandler(log *slog.Logger, timeout time.Duration, normalizer core.Normalizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		phrase := strings.TrimSpace(r.URL.Query().Get("phrase"))
		if phrase == "" {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: core.ErrBadArguments.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		words, err := normalizer.Norm(ctx, phrase)
		cancel()

		if err != nil {
			if status.Code(err) == codes.ResourceExhausted {
				writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
				return
			}

			log.Error("norm failed", "error", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
			return
		}

		writeJSON(w, http.StatusOK, wordsResponse{
			Words: words,
			Total: len(words),
		})
	}
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

var _ = errors.New
