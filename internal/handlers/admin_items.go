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
	"github.com/gorilla/csrf"
	"github.com/nfnt/resize"
	"github.com/google/uuid"
)

func (h *AdminHandler) EditItemForm(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	item, err := h.Store.GetItemByID(id)
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	tmpl := h.Templates.Get("admin_edit_item.html")
	if tmpl == nil {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}
	session, _ := h.SessionStore.Get(r, "admin-session")
	data := map[string]interface{}{
		"CsrfField": csrf.TemplateField(r),
		"Flashes":   GetFlash(session),
		"Item":      item,
	}
	session.Save(r, w)
	tmpl.Execute(w, data)
}

func (h *AdminHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	session, _ := h.SessionStore.Get(r, "admin-session")
	defer session.Save(r, w)

	err := r.ParseMultipartForm(10 << 20) // 10MB
	if err != nil {
		session.AddFlash(FlashMessage{Type: "error", Message: "File too large."})
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	idStr := r.FormValue("id")
	id, _ := strconv.Atoi(idStr)
	title := r.FormValue("title")
	desc := r.FormValue("description")
	priceStr := r.FormValue("price")
	delivery := r.FormValue("delivery_time")
	status := r.FormValue("status")

	price, _ := strconv.ParseFloat(priceStr, 64)

	item := &models.Item{
		ID:           id,
		Title:        title,
		Description:  desc,
		Price:        price,
		DeliveryTime: delivery,
		Status:       status,
	}

	if err := h.Store.UpdateItem(item); err != nil {
		session.AddFlash(FlashMessage{Type: "error", Message: "Error updating item."})
		http.Redirect(w, r, fmt.Sprintf("/admin/items/edit?id=%d", id), http.StatusSeeOther)
		return
	}

	// Handle optional image update
	file, header, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		
		var img image.Image
		ext := filepath.Ext(header.Filename)
		if ext == ".png" {
			img, err = png.Decode(file)
		} else if ext == ".jpeg" || ext == ".jpg" {
			img, err = jpeg.Decode(file)
		} else {
			session.AddFlash(FlashMessage{Type: "error", Message: "Unsupported image format."})
			http.Redirect(w, r, fmt.Sprintf("/admin/items/edit?id=%d", id), http.StatusSeeOther)
			return
		}

		if err == nil {
			newImage := resize.Resize(800, 0, img, resize.Lanczos3)
			filename := fmt.Sprintf("%s.jpg", uuid.New().String())
			uploadPath := filepath.Join("static/uploads", filename)

			out, _ := os.Create(uploadPath)
			jpeg.Encode(out, newImage, &jpeg.Options{Quality: 80})
			out.Close()

			h.Store.UpdateItemImage(id, "/static/uploads/"+filename)
		}
	}

	session.AddFlash(FlashMessage{Type: "success", Message: "Item updated successfully!"})
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
