package handlers

import (
	"net/http"
	"strconv"

	"github.com/gorilla/csrf"
)

func (h *AdminHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limitStr := r.URL.Query().Get("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10 // Default limit
	}

	offset := (page - 1) * limit

	orders, err := h.Store.GetAllOrders(limit, offset)
	if err != nil {
		http.Error(w, "Error fetching orders", http.StatusInternalServerError)
		return
	}

	totalOrders, err := h.Store.GetTotalOrdersCount()
	if err != nil {
		http.Error(w, "Error fetching total order count", http.StatusInternalServerError)
		return
	}

	totalPages := (totalOrders + limit - 1) / limit
	if totalPages == 0 { // Handle case with no orders
		totalPages = 1
	}

	tmpl := h.Templates.Get("admin_orders.html")
	if tmpl == nil {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	session, _ := h.SessionStore.Get(r, "admin-session")
	data := map[string]interface{}{
		"Orders":    orders,
		"CsrfField": csrf.TemplateField(r),
		"Flashes":   GetFlash(session),
		"CurrentPage": page,
		"TotalPages":  totalPages,
		"Limit":       limit,
	}
	session.Save(r, w)
	tmpl.Execute(w, data)
}

func (h *AdminHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	idStr := r.FormValue("id")
	status := r.FormValue("status")
	adminComments := r.FormValue("admin_comments")
	
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.Store.UpdateOrderStatus(id, status, adminComments); err != nil {
		http.Error(w, "Error updating status", http.StatusInternalServerError)
		return
	}

	session, _ := h.SessionStore.Get(r, "admin-session")
	session.AddFlash(FlashMessage{Type: "success", Message: "Order updated!"})
	session.Save(r, w)
	http.Redirect(w, r, "/admin/orders", http.StatusSeeOther)
}
