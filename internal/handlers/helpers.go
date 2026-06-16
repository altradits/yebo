package handlers

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

var tmplCache map[string]*template.Template

// InitTemplates parses and caches all templates from the given directory.
func InitTemplates(dir string) error {
	tmplCache = make(map[string]*template.Template)
	layout := filepath.Join(dir, "layout.html")
	patterns := []string{
		"*.html",
		"customer/*.html",
		"agent/*.html",
		"trader/*.html",
		"admin/*.html",
	}
	for _, pattern := range patterns {
		files, err := filepath.Glob(filepath.Join(dir, pattern))
		if err != nil {
			return err
		}
		for _, f := range files {
			name, err := filepath.Rel(dir, f)
			if err != nil {
				return err
			}
			if filepath.Base(f) == "layout.html" || filepath.Base(f) == "home.html" {
				continue
			}
			t, err := template.ParseFiles(layout, f)
			if err != nil {
				return err
			}
			tmplCache[name] = t
		}
	}
	// home.html and login/register are standalone (no layout)
	for _, standalone := range []string{"home.html", "login.html", "register.html"} {
		t, err := template.ParseFiles(filepath.Join(dir, standalone))
		if err != nil {
			return err
		}
		tmplCache[standalone] = t
	}
	return nil
}

func renderTemplate(w http.ResponseWriter, r *http.Request, name string, data interface{}) {
	t, ok := tmplCache[name]
	if !ok {
		log.Printf("template not found: %s", name)
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.Execute(w, data); err != nil {
		log.Printf("template execute %s: %v", name, err)
	}
}
