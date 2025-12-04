package handlers

import (
	"net/http"

	"github.com/gorilla/csrf"
)

func (h *AdminHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	// Admin sees ALL items including archived
	items, err := h.Store.GetAllItems()
	if err != nil {
		http.Error(w, "Error fetching items", http.StatusInternalServerError)
		return
	}

	tmpl := h.Templates.Get("admin_items.html")
	if tmpl == nil {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}
	session, _ := h.SessionStore.Get(r, "admin-session")
	data := map[string]interface{}{
		"Items":     items,
		"Flashes":   GetFlash(session),
		"CsrfField": csrf.TemplateField(r),
	}
	session.Save(r, w)
	tmpl.Execute(w, data)
}
