package handlers

import (
	"encoding/gob"
	"log/slog"
	"net/http"
	"sync"
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

// RateLimiter struct to hold state
type RateLimiter struct {
	visitors sync.Map
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter with a cleanup goroutine
func NewRateLimiter(window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		window: window,
	}
	// Background cleanup
	go rl.cleanup()
	return rl
}

// cleanup removes old entries to prevent memory leaks
func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(time.Minute)
		now := time.Now()
		rl.visitors.Range(func(key, value interface{}) bool {
			lastSeen := value.(time.Time)
			if now.Sub(lastSeen) > rl.window {
				rl.visitors.Delete(key)
			}
			return true
		})
	}
}

// Middleware enforces the rate limit
func (rl *RateLimiter) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only rate limit POST requests or specific actions if needed, 
		// but for this general middleware, we assume it's applied to specific routes.
		ip := r.RemoteAddr 

		if lastSeen, ok := rl.visitors.Load(ip); ok {
			if time.Since(lastSeen.(time.Time)) < rl.window {
				slog.Warn("Rate limit exceeded", "ip", ip)
				// Optionally add a Flash message if sessions are available/wanted here, 
				// but a simple error is standard for 429.
				http.Error(w, "Too Many Requests. Please try again later.", http.StatusTooManyRequests)
				return
			}
		}

		rl.visitors.Store(ip, time.Now())
		next(w, r)
	}
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