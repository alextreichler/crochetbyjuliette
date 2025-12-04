package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alextreichler/crochetbyjuliette/internal/config"
	"github.com/alextreichler/crochetbyjuliette/internal/handlers"
	"github.com/alextreichler/crochetbyjuliette/internal/store"
	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
)

func main() {
	// Configure slog to output DEBUG level messages
	// This should be done as early as possible in main
	handlerOpts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	// Using TextHandler for console readability; for production JSONHandler might be preferred.
	logger := slog.New(slog.NewTextHandler(os.Stdout, handlerOpts))
	slog.SetDefault(logger)

	// 1. Load Configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// 2. Init DB
	db, err := store.NewStore(cfg.DBPath)
	if err != nil {
		slog.Error("Failed to initialize store", "error", err)
		os.Exit(1)
	}

	// Run Migrations
	if err := db.Migrate("migrations"); err != nil {
		slog.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}

	// 3. Session Setup
	sessionStore := sessions.NewCookieStore(cfg.SessionKey)
	sessionStore.Options.HttpOnly = true
	sessionStore.Options.Secure = cfg.CookieSecure // Configurable for production
	sessionStore.Options.SameSite = http.SameSiteLaxMode
	sessionStore.Options.Path = "/"
	if cfg.CookieDomain != "" {
		sessionStore.Options.Domain = cfg.CookieDomain
	}

	// 3. Init Templates
	templates := handlers.NewTemplateCache()

	// Add other template funcs (prevPage, nextPage)
	templates.AddFunc("prevPage", func(currentPage int) int { return currentPage - 1 })
	templates.AddFunc("nextPage", func(currentPage int) int { return currentPage + 1 })

	if err := templates.Load("templates"); err != nil {
		slog.Error("Failed to load templates", "error", err)
		os.Exit(1)
	}

	// 4. Setup Handlers
	adminHandler := &handlers.AdminHandler{
		Store:        db,
		SessionStore: sessionStore,
		Templates:    templates,
	}
	homeHandler := &handlers.HomeHandler{
		Store:        db,
		Templates:    templates,
		SessionStore: sessionStore,
	}
	orderHandler := &handlers.OrderHandler{
		Store:        db,
		Templates:    templates,
		SessionStore: sessionStore,
	}
	mux := http.NewServeMux()

	// Static Files
	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	// Rate Limiter (1 request per minute)
	rateLimiter := handlers.NewRateLimiter(1 * time.Minute)

	// Public Routes
	mux.HandleFunc("/", homeHandler.Index)
	mux.HandleFunc("/order", orderHandler.OrderForm)                                // GET form
	mux.HandleFunc("POST /order", rateLimiter.Middleware(orderHandler.SubmitOrder)) // POST submit

	// Order Status (Magic Link)
	mux.HandleFunc("/status-request", orderHandler.RequestStatusLink) // GET form & POST submit (could split)
	mux.HandleFunc("POST /status-request", rateLimiter.Middleware(orderHandler.SendStatusLink))
	mux.HandleFunc("/my-orders", orderHandler.MyOrders)            // List all orders for valid token
	mux.HandleFunc("/order/status/", orderHandler.ViewOrderStatus) // Trailing slash matches /order/status/{token}

	// Order Management (Edit/Cancel)
	mux.HandleFunc("/order/edit/", orderHandler.EditOrderForm)
	mux.HandleFunc("POST /order/update", rateLimiter.Middleware(orderHandler.UpdateOrder))
	mux.HandleFunc("POST /order/cancel", rateLimiter.Middleware(orderHandler.CancelOrder))

	mux.HandleFunc("/login", adminHandler.LoginGet)
	mux.HandleFunc("POST /login", adminHandler.LoginPost)
	mux.HandleFunc("/logout", adminHandler.Logout)

	// Protected Routes
	mux.HandleFunc("/admin", adminHandler.AuthMiddleware(adminHandler.Dashboard))
	mux.HandleFunc("/admin/orders", adminHandler.AuthMiddleware(adminHandler.ListOrders))
	mux.HandleFunc("POST /admin/orders/update", adminHandler.AuthMiddleware(adminHandler.UpdateOrderStatus))

	mux.HandleFunc("/admin/items", adminHandler.AuthMiddleware(adminHandler.ListItems))       // List all items
	mux.HandleFunc("/admin/items/new", adminHandler.AuthMiddleware(adminHandler.AddItemForm)) // GET form
	mux.HandleFunc("POST /admin/items", adminHandler.AuthMiddleware(adminHandler.CreateItem)) // POST submit
	mux.HandleFunc("POST /admin/items/delete", adminHandler.AuthMiddleware(adminHandler.DeleteItem))
	mux.HandleFunc("/admin/items/edit", adminHandler.AuthMiddleware(adminHandler.EditItemForm))      // GET form
	mux.HandleFunc("POST /admin/items/update", adminHandler.AuthMiddleware(adminHandler.UpdateItem)) // POST submit
	// 6. Middleware Setup
	CSRF := csrf.Protect(
		cfg.CSRFKey,
		csrf.Secure(cfg.CookieSecure), // Configurable for production
		// Fix for "Forbidden - origin invalid": Trust local development origins
		csrf.TrustedOrigins([]string{"localhost:" + cfg.Port, "127.0.0.1:" + cfg.Port, "localhost", "127.0.0.1"}),
	)

	// Wrap the router with middleware chain
	// Chain: Logger -> Security Headers -> CSRF -> Mux
	handler := handlers.LoggingMiddleware(
		handlers.SecurityHeadersMiddleware(
			CSRF(mux),
		),
	)

	// 7. Start Server with Graceful Shutdown
	server := &http.Server{
		Addr:    ":" + cfg.Port, // Use ENV var, default 8585 already set in ENV
		Handler: handler,
	}

	// Create a channel to listen for OS signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Goroutine to start the server
	go func() {
		slog.Info("Server starting", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed to listen and serve", "error", err)
			os.Exit(1)
		}
	}()

	// Block until a signal is received
	<-stop

	slog.Info("Shutting down server gracefully...")

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("Server exited gracefully.")
}
