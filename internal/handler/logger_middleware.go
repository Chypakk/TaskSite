package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"tasksite/internal/logger"
	"time"
)

type contextKey string

const loggerKey contextKey = "logger"

func WithLogger(ctx context.Context, logger *logger.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func FromContext(ctx context.Context) *logger.Logger {
	if l, ok := ctx.Value(loggerKey).(*logger.Logger); ok {
		return l
	}
	return logger.New()
}

func RequestLogger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqID := generateShortID()
		start := time.Now()

		log := logger.New()

		ctx := context.WithValue(r.Context(), "request_id", reqID)
		ctx = WithLogger(ctx, log)

		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next(rw, r.WithContext(ctx))

		log.Info(ctx, "[%s] %s %s → %d (%v)",
			reqID, r.Method, r.URL.Path, rw.statusCode, time.Since(start))
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
func (rw *responseWriter) Header() http.Header {
	rw.ResponseWriter.Header().Set("X-Request-Id",
		rw.ResponseWriter.Header().Get("X-Request-Id"))
	return rw.ResponseWriter.Header()
}

func generateShortID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
