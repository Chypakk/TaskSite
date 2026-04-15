package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"tasksite/internal/logger"
	"time"
)

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := generateShortID()
		start := time.Now()

		log := logger.New()

		ctx := context.WithValue(r.Context(), "request_id", reqID)
		ctx = logger.WithLogger(ctx, log)

		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r.WithContext(ctx))

		log.Info(ctx, "request_completed",
			"req_id", reqID,
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.statusCode,
			"duration", time.Since(start))
	})
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
