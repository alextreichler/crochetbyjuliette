package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/alextreichler/crochetbyjuliette/internal/models"
	"github.com/alextreichler/crochetbyjuliette/internal/store"
	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
)

type OrderHandler struct {
	Store        *store.Store
	Templates    *TemplateCache
	SessionStore *sessions.CookieStore
}

func (h *OrderHandler) OrderForm(w http.ResponseWriter, r *http.Request) {
	// Parse ID from URL query for simplicity, or could use path params if we had a router
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Item ID", http.StatusBadRequest)
		return
	}

	item, err := h.Store.GetItemByID(id)
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	tmpl := h.Templates.Get("order.html")
	if tmpl == nil {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}
	session, _ := h.SessionStore.Get(r, "order-session")
	data := map[string]interface{}{
		"Item":      item,
		"CsrfField": csrf.TemplateField(r),
		"Flashes":   GetFlash(session),
	}
	session.Save(r, w)
	tmpl.Execute(w, data)
}

func generateToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "fallback-token-" + strconv.FormatInt(time.Now().Unix(), 10)
	}
	return hex.EncodeToString(b)
}

func generateOrderRef() string {
	// Generate 8 chars alphanumeric (uppercase)
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Removed I, O, 1, 0 to avoid confusion
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "ORD" + strconv.FormatInt(time.Now().Unix(), 10)
	}
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}

func (h *OrderHandler) SubmitOrder(w http.ResponseWriter, r *http.Request) {
	session, _ := h.SessionStore.Get(r, "order-session") // Using a different session store for public orders
	defer session.Save(r, w)

	if err := r.ParseForm(); err != nil {
		session.AddFlash(FlashMessage{Type: "error", Message: "Invalid form data."})
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther) // Redirect back to form
		return
	}

	itemID, err := strconv.Atoi(r.FormValue("item_id"))
	if err != nil {
		session.AddFlash(FlashMessage{Type: "error", Message: "Invalid item ID."})
		http.Redirect(w, r, "/", http.StatusSeeOther) // Redirect to home
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	address := r.FormValue("address")
	notes := r.FormValue("notes")
	qtyStr := r.FormValue("quantity")
	quantity := 1
	if qtyStr != "" {
		if q, err := strconv.Atoi(qtyStr); err == nil && q > 0 {
			quantity = q
		}
	}

	// Validation
	errors := make(map[string]string)
	if name == "" {
		errors["name"] = "Your name is required."
	}
	if email == "" {
		errors["email"] = "Email address is required."
	} else if !isValidEmail(email) {
		errors["email"] = "Please enter a valid email address."
	}
	if address == "" {
		errors["address"] = "Shipping address is required."
	}

	if len(errors) > 0 {
		for _, msg := range errors {
			session.AddFlash(FlashMessage{Type: "error", Message: msg})
		}
		// Redirect back to form, preserving values if possible (not implemented here yet)
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther) // Redirect back to form
		return
	}

	token := generateToken()
	orderRef := generateOrderRef()

	order := &models.Order{
		ItemID:          itemID,
		OrderRef:        orderRef,
		Quantity:        quantity,
		CustomerName:    name,
		CustomerEmail:   email,
		CustomerAddress: address,
		Status:          "Ordered",
		Notes:           notes,
		MagicToken:      token,
		MagicTokenExpiry: time.Now().Add(30 * 24 * time.Hour),
	}

	if err := h.Store.CreateOrder(order); err != nil {
		session.AddFlash(FlashMessage{Type: "error", Message: "Failed to place order. Please try again."})
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther) // Redirect back to form
		return
	}

	// MOCK EMAIL SENDING
	slog.Info("==========================================")
	slog.Info("📧 EMAIL SENT TO: " + email)
	slog.Info("Subject: Order Confirmation - Crochet by Juliette")
	slog.Info("Order Reference: " + orderRef)
	slog.Info("Your Magic Link: http://localhost:8585/order/status/" + token)
	slog.Info("==========================================")

	session.AddFlash(FlashMessage{Type: "success", Message: "Order placed successfully! Check your email for details."})
	// Redirect directly to the Order Status page (Magic Link)
	http.Redirect(w, r, "/order/status/"+token, http.StatusSeeOther)
}

// Basic email validation regex
var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}