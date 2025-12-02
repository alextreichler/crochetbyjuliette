package handlers

import (
	"encoding/gob"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
)

// Register types for gob encoding (used by sessions)
func init() {
	gob.Register(FlashMessage{})
}

// LoggingMiddleware logs the details of each HTTP request
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// Wrap ResponseWriter to capture status code
		ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(ww, r)
		slog.Info("HTTP Request",
			"method", r.Method,
			"path", r.URL.Path, // Use r.URL.Path instead of r.PathValue("") which is for specific route matches
			"status", ww.statusCode,
			"duration", time.Since(start),
			"ip", r.RemoteAddr,
		)
	})
}

// Custom ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// SecurityHeadersMiddleware adds standard security headers
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		// CSP: Allow styles/scripts from self and standard sources.
		// Note: Inline styles might break with strict CSP unless 'unsafe-inline' is used (common compromise for simple apps)
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; script-src 'self'") // Removed 'unsafe-eval' to tighten CSP.
		next.ServeHTTP(w, r)
	})
}

// FlashMessage structure
type FlashMessage struct {
	Type    string
	Message string
}

// GetFlash retrieves flash messages from the session
func GetFlash(session *sessions.Session) []FlashMessage {
	flashes := session.Flashes()
	var messages []FlashMessage
	for _, f := range flashes {
		if fm, ok := f.(FlashMessage); ok {
			messages = append(messages, fm)
		}
	}
	return messages
}
