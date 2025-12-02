package handlers

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/csrf"
)

func (h *OrderHandler) RequestStatusLink(w http.ResponseWriter, r *http.Request) {
	tmpl := h.Templates.Get("status_request.html")
	if tmpl == nil {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}
	session, _ := h.SessionStore.Get(r, "order-session")
	data := map[string]interface{}{
		"CsrfField": csrf.TemplateField(r),
		"Flashes":   GetFlash(session),
	}
	session.Save(r, w) // Save session to clear flashes
	tmpl.Execute(w, data)
}

func (h *OrderHandler) SendStatusLink(w http.ResponseWriter, r *http.Request) {
	session, _ := h.SessionStore.Get(r, "order-session")
	defer session.Save(r, w)

	email := r.FormValue("email")
	
	// Check if any orders exist for this email (case-insensitive)
	orders, err := h.Store.GetOrdersByEmail(email)
	if err != nil {
		session.AddFlash(FlashMessage{Type: "error", Message: "Internal Error processing your request."})
		http.Redirect(w, r, "/status-request", http.StatusInternalServerError) // Use Redirect instead of http.Error
		return
	}

	if len(orders) > 0 {
		// Generate a global access token for "My Orders" page
		token := generateToken()
		
		if err := h.Store.CreateLoginToken(email, token); err != nil {
			session.AddFlash(FlashMessage{Type: "error", Message: "Error generating access link. Please try again."})
			http.Redirect(w, r, "/status-request", http.StatusInternalServerError)
			return
		}

		// MOCK EMAIL
		slog.Info("==========================================")
		slog.Info("📧 EMAIL SENT TO: " + email)
		slog.Info("Subject: Your Orders - Crochet by Juliette")
		slog.Info("Access All Orders: http://localhost:8585/my-orders?token=" + token)
		slog.Info("==========================================")
	} else {
		// Security: Don't reveal if email exists, but maybe log it.
		slog.Info("Status requested for unknown email: " + email)
	}

	// Show "Check your email" message regardless of success (security)
	session.AddFlash(FlashMessage{Type: "success", Message: "If you have active orders, a link has been sent to your email."})
	http.Redirect(w, r, "/status-request", http.StatusSeeOther)
}

func (h *OrderHandler) MyOrders(w http.ResponseWriter, r *http.Request) {
	session, _ := h.SessionStore.Get(r, "order-session")
	defer session.Save(r, w)

	token := r.URL.Query().Get("token")
	if token == "" {
		session.AddFlash(FlashMessage{Type: "error", Message: "Missing access token."})
		http.Redirect(w, r, "/status-request", http.StatusSeeOther)
		return
	}

	// Validate token and get email
	email, err := h.Store.GetEmailByLoginToken(token)
	if err != nil || email == "" {
		session.AddFlash(FlashMessage{Type: "error", Message: "Invalid or Expired Link. Please request a new one."})
		http.Redirect(w, r, "/status-request", http.StatusSeeOther)
		return
	}

	// Fetch all orders for this verified email
	orders, err := h.Store.GetOrdersByEmail(email)
	if err != nil {
		session.AddFlash(FlashMessage{Type: "error", Message: "Error fetching your orders."})
		http.Redirect(w, r, "/status-request", http.StatusSeeOther)
		return
	}

	tmpl := h.Templates.Get("my_orders.html")
	if tmpl == nil {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, map[string]interface{}{
		"Orders":  orders,
		"Email":   email,
		"Flashes": GetFlash(session),
	})
}

func (h *OrderHandler) ViewOrderStatus(w http.ResponseWriter, r *http.Request) {
	session, _ := h.SessionStore.Get(r, "order-session")
	defer session.Save(r, w)

	// Extract token from path manually since we use ServeMux
	// Path is /order/status/{token}
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		session.AddFlash(FlashMessage{Type: "error", Message: "Invalid order link."})
		http.Redirect(w, r, "/status-request", http.StatusSeeOther)
		return
	}
	token := parts[3]

	order, err := h.Store.GetOrderByToken(token)
	if err != nil {
		session.AddFlash(FlashMessage{Type: "error", Message: "Order not found or link is invalid."})
		http.Redirect(w, r, "/status-request", http.StatusSeeOther)
		return
	}

	// Check expiry
	if time.Now().After(order.MagicTokenExpiry) {
		session.AddFlash(FlashMessage{Type: "error", Message: "Link Expired. Please request a new one."})
		http.Redirect(w, r, "/status-request", http.StatusSeeOther)
		return
	}

	tmpl := h.Templates.Get("order_status.html")
	if tmpl == nil {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}
	data := map[string]interface{}{
		"Order":   order,
		"Flashes": GetFlash(session),
	}
	session.Save(r, w) // Save after getting flashes to clear them
	tmpl.Execute(w, data)
}