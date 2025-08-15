package middleware

import (
	"net/http"
	"time"

	"github.com/Osminalx/fluxio/pkg/utils/logger"
)

// LoggingMiddleware logs all HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		start := time.Now()
		
		// Create a custom response writer to capture status code
		responseWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		// Call the next handler
		next.ServeHTTP(responseWriter, r)
		
		// Calculate duration
		duration := time.Since(start)
		
		// Log the request
		logger.HTTPRequest(
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
			responseWriter.statusCode,
			duration,
			r.UserAgent(),
		)
		
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}
