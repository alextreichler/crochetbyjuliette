package handlers

import (
	"html/template"
	"log/slog"
	"path/filepath"
	"sync"
)

// TemplateCache holds parsed templates
type TemplateCache struct {
	cache map[string]*template.Template
	mu    sync.RWMutex
	funcs template.FuncMap
}

func NewTemplateCache() *TemplateCache {
	return &TemplateCache{
		cache: make(map[string]*template.Template),
		funcs: make(template.FuncMap),
	}
}

func (tc *TemplateCache) AddFunc(name string, fn interface{}) {
			tc.mu.Lock()
		defer tc.mu.Unlock()
		tc.funcs[name] = fn
	}
	
	// Load parses all templates in the templates/ dir
	func (tc *TemplateCache) Load(dir string) error {
		tc.mu.Lock()
		defer tc.mu.Unlock()
	
		// Add global template functions
		tc.funcs["prevPage"] = func(currentPage int) int {
			return currentPage - 1
		}
		tc.funcs["nextPage"] = func(currentPage int) int {
			return currentPage + 1
		}
	
		// Find all HTML files
		files, err := filepath.Glob(filepath.Join(dir, "*.html"))
		if err != nil {
			return err
		}
	for _, file := range files {
		name := filepath.Base(file)
		tmpl, err := template.New(name).Funcs(tc.funcs).ParseFiles(file)
		if err != nil {
			slog.Error("Failed to parse template", "file", file, "error", err)
			return err
		}
		tc.cache[name] = tmpl
		slog.Debug("Cached template", "name", name)
	}
	return nil
}

func (tc *TemplateCache) Get(name string) *template.Template {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.cache[name]
}
