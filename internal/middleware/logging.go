package middleware

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(data)
	rw.written += int64(n)
	return n, err
}

// LoggingMiddleware provides structured HTTP request logging
func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response wrapper to capture metrics
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Process the request
		next.ServeHTTP(rw, r)

		// Log the request with structured fields
		duration := time.Since(start)

		logEntry := logrus.WithFields(logrus.Fields{
			"method":        r.Method,
			"path":          r.URL.Path,
			"query":         r.URL.RawQuery,
			"status_code":   rw.statusCode,
			"duration_ms":   duration.Milliseconds(),
			"bytes_written": rw.written,
			"remote_addr":   getClientIP(r),
			"user_agent":    r.UserAgent(),
			"referer":       r.Referer(),
		})

		// Log at appropriate level based on status code
		switch {
		case rw.statusCode >= 500:
			logEntry.Error("HTTP request completed with server error")
		case rw.statusCode >= 400:
			logEntry.Warn("HTTP request completed with client error")
		case rw.statusCode >= 300:
			logEntry.Info("HTTP request completed with redirect")
		default:
			logEntry.Info("HTTP request completed successfully")
		}
	})
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check for forwarded headers first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	// Fall back to remote address
	return r.RemoteAddr
}
