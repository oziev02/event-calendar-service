package http

import (
	"net/http"
	"time"

	"github.com/oziev02/event-calendar-service/pkg/logger"
)

// LoggingMiddleware логирует HTTP запросы
type LoggingMiddleware struct {
	logger logger.Logger
}

// NewLoggingMiddleware создает новый middleware для логирования
func NewLoggingMiddleware(log logger.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{logger: log}
}

// Handler оборачивает HTTP обработчик с логированием
func (m *LoggingMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Создать обертку response writer для захвата статус кода
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		m.logger.Log(logger.LevelInfo, "HTTP Request", map[string]interface{}{
			"method":      r.Method,
			"url":         r.URL.String(),
			"status_code": rw.statusCode,
			"duration_ms": duration.Milliseconds(),
		})
	})
}

// responseWriter оборачивает http.ResponseWriter для захвата статус кода
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
