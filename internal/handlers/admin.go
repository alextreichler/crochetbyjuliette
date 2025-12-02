package handlers

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/alextreichler/crochetbyjuliette/internal/models"
	"github.com/alextreichler/crochetbyjuliette/internal/store"
	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
	"github.com/nfnt/resize"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"github.com/google/uuid"
)

type AdminHandler struct {
	Store        *store.Store
	SessionStore *sessions.CookieStore
	Templates    *TemplateCache
}

func (h *AdminHandler) LoginGet(w http.ResponseWriter, r *http.Request) {
	tmpl := h.Templates.Get("login.html")
	if tmpl == nil {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}
	session, _ := h.SessionStore.Get(r, "admin-session")
	data := map[string]interface{}{
		"CsrfField": csrf.TemplateField(r),
		"Flashes":   GetFlash(session),
	}
	session.Save(r, w)
	tmpl.Execute(w, data)
}

func (h *AdminHandler) LoginPost(w http.ResponseWriter, r *http.Request) {
	session, _ := h.SessionStore.Get(r, "admin-session")

	username := r.FormValue("username")
	password := r.FormValue("password")

	user, err := h.Store.GetUserByUsername(username)
	if err != nil {
		session.AddFlash(FlashMessage{Type: "error", Message: "Internal Server Error"})
		session.Save(r, w) // Save before redirect
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if user == nil {
		session.AddFlash(FlashMessage{Type: "error", Message: "Invalid username or password"})
		session.Save(r, w) // Save before redirect
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		session.AddFlash(FlashMessage{Type: "error", Message: "Invalid username or password"})
		session.Save(r, w) // Save before redirect
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Set authenticated session
	session.Values["authenticated"] = true
	session.Values["user_id"] = user.ID
	session.Options.Path = "/" // Ensure the cookie is valid for all paths
	session.AddFlash(FlashMessage{Type: "success", Message: "Welcome, " + user.Username + "!"})

	// CRITICAL: Save session and check for errors
	if err := session.Save(r, w); err != nil {
		slog.Error("Failed to save session", "error", err)
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	slog.Info("Login successful, redirecting to /admin", "user_id", user.ID)
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *AdminHandler) Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := h.SessionStore.Get(r, "admin-session")
	session.Values["authenticated"] = false
	session.Options.MaxAge = -1 // Expire immediately
	session.AddFlash(FlashMessage{Type: "success", Message: "Logged out successfully!"})
	session.Save(r, w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// AuthMiddleware ensures the user is logged in
func (h *AdminHandler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("AuthMiddleware triggered for path", "path", r.URL.Path)
		session, _ := h.SessionStore.Get(r, "admin-session")
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			slog.Info("AuthMiddleware: User not authenticated, redirecting to /login", "path", r.URL.Path)
			session.AddFlash(FlashMessage{Type: "error", Message: "You must be logged in to access this page."})
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		slog.Info("AuthMiddleware: User authenticated", "user_id", session.Values["user_id"], "path", r.URL.Path)
		next(w, r)
	}
}

// New methods moved from dashboard.go
func (h *AdminHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	stats, err := h.Store.GetDashboardStats()
	if err != nil {
		http.Error(w, "Error fetching stats", http.StatusInternalServerError)
		return
	}

	tmpl := h.Templates.Get("admin.html")
	if tmpl == nil {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}
	// Add flash messages to the data
	session, _ := h.SessionStore.Get(r, "admin-session")
	data := map[string]interface{}{
		"Stats":   stats,
		"Flashes": GetFlash(session),
	}
	session.Save(r, w) // Save session to clear flashes
	tmpl.Execute(w, data)
}

func (h *AdminHandler) AddItemForm(w http.ResponseWriter, r *http.Request) {
	tmpl := h.Templates.Get("admin_add_item.html")
	if tmpl == nil {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}
	session, _ := h.SessionStore.Get(r, "admin-session")
	data := map[string]interface{}{
		"CsrfField": csrf.TemplateField(r),
		"Flashes":   GetFlash(session),
		"Values":    r.Form, // Pre-fill form on error
	}
	session.Save(r, w) // Save session to clear flashes
	tmpl.Execute(w, data)
}

func (h *AdminHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	session, _ := h.SessionStore.Get(r, "admin-session")
	defer session.Save(r, w) // Save session to clear flashes if any were added and to persist others

	// 1. Parse Multipart Form
	err := r.ParseMultipartForm(10 << 20) // 10MB
	if err != nil {
		session.AddFlash(FlashMessage{Type: "error", Message: "File too large. Max 10MB."})
		http.Redirect(w, r, "/admin/items/new", http.StatusSeeOther)
		return
	}

	title := r.FormValue("title")
	desc := r.FormValue("description")
	priceStr := r.FormValue("price")
	delivery := r.FormValue("delivery_time")
	status := r.FormValue("status")
	if status == "" {
		status = "available"
	}

	// Validation
	errors := make(map[string]string)
	if title == "" {
		errors["title"] = "Title is required."
	}
	if priceStr == "" {
		errors["price"] = "Price is required."
	}
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		errors["price"] = "Invalid price format."
	} else if price <= 0 {
		errors["price"] = "Price must be positive."
	}
	if delivery == "" {
		errors["delivery"] = "Delivery time is required."
	}
	validStatuses := map[string]bool{"available": true, "out_of_stock": true, "archived": true}
	if !validStatuses[status] {
		errors["status"] = "Invalid status selected."
	}

	file, header, fileErr := r.FormFile("image")
	if fileErr != nil {
		errors["image"] = "Image file is required."
	}

	if len(errors) > 0 {
		for _, msg := range errors {
			session.AddFlash(FlashMessage{Type: "error", Message: msg})
		}
		// Redirect back to form, preserving values if possible (r.Form already has them)
		http.Redirect(w, r, "/admin/items/new", http.StatusSeeOther)
		return
	}
	// If fileErr was nil, it means file exists, defer close.
	defer file.Close()

	// 2. Handle File Upload and Optimization
	// Decode image
	var img image.Image
	ext := filepath.Ext(header.Filename)
	if ext == ".png" {
		img, err = png.Decode(file)
	} else if ext == ".jpeg" || ext == ".jpg" { // Explicitly handle JPEG
		img, err = jpeg.Decode(file)
	} else {
		session.AddFlash(FlashMessage{Type: "error", Message: "Unsupported image format. Only PNG, JPG, JPEG are allowed."})
		http.Redirect(w, r, "/admin/items/new", http.StatusSeeOther)
		return
	}

	if err != nil {
		session.AddFlash(FlashMessage{Type: "error", Message: "Failed to decode image."})
		http.Redirect(w, r, "/admin/items/new", http.StatusSeeOther)
		return
	}

	// Resize image (max width 800px, preserve aspect ratio)
	newImage := resize.Resize(800, 0, img, resize.Lanczos3)

	// Create a unique filename
	filename := fmt.Sprintf("%s.jpg", uuid.New().String()) // Use UUID for unique filenames
	uploadPath := filepath.Join("static/uploads", filename)

	// Save file to disk
	out, err := os.Create(uploadPath)
	if err != nil {
		session.AddFlash(FlashMessage{Type: "error", Message: "Error saving image file."})
		http.Redirect(w, r, "/admin/items/new", http.StatusSeeOther)
		return
	}
	defer out.Close()

	// Write new image to file
	err = jpeg.Encode(out, newImage, &jpeg.Options{Quality: 80}) // Add quality for JPEG
	if err != nil {
		session.AddFlash(FlashMessage{Type: "error", Message: "Error encoding image."})
		http.Redirect(w, r, "/admin/items/new", http.StatusSeeOther)
		return
	}

	// 3. Create Item in DB
	item := &models.Item{
		Title:        title,
		Description:  desc,
		Price:        price,
		DeliveryTime: delivery,
		ImageURL:     "/static/uploads/" + filename,
		Status:       status,
	}

	if err := h.Store.CreateItem(item); err != nil {
		session.AddFlash(FlashMessage{Type: "error", Message: "Error saving item to database."})
		http.Redirect(w, r, "/admin/items/new", http.StatusSeeOther)
		return
	}

	session.AddFlash(FlashMessage{Type: "success", Message: "Item added successfully!"})
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *AdminHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	session, _ := h.SessionStore.Get(r, "admin-session")
	defer session.Save(r, w)

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		session.AddFlash(FlashMessage{Type: "error", Message: "Invalid ID."})
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	if err := h.Store.DeleteItem(id); err != nil {
		session.AddFlash(FlashMessage{Type: "error", Message: "Error deleting item."})
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	session.AddFlash(FlashMessage{Type: "success", Message: "Item deleted successfully!"})
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
