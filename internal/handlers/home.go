package handlers

import (
	"net/http"

	"github.com/alextreichler/crochetbyjuliette/internal/store"
	"github.com/gorilla/sessions"
)

type HomeHandler struct {
	Store     *store.Store
	Templates *TemplateCache
	SessionStore *sessions.CookieStore
}

func (h *HomeHandler) Index(w http.ResponseWriter, r *http.Request) {
	items, err := h.Store.GetPublicItems()
	if err != nil {
		http.Error(w, "Error fetching items", http.StatusInternalServerError)
		return
	}

	tmpl := h.Templates.Get("home.html")
	if tmpl == nil {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}
	publicSession, _ := h.SessionStore.Get(r, "public-session")
	adminSession, _ := h.SessionStore.Get(r, "admin-session")

	isAdmin := false
	if auth, ok := adminSession.Values["authenticated"].(bool); ok && auth {
		isAdmin = true
	}

	data := map[string]interface{}{
		"Items":   items,
		"Flashes": GetFlash(publicSession),
		"IsAdmin": isAdmin,
	}
	publicSession.Save(r, w)
	tmpl.Execute(w, data)
}