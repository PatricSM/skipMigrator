package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/PatricSM/skip-migrator/backend/internal/db"
)

type handlers struct{ deps Deps }

// POST /api/migrations
//
// multipart/form-data:
//
//	source: zip file
//	pixel_perfect: "true"|"false"
//	validate: "true"|"false"
//	supabase_strategy: "extract"|"new"
//
// Response: { "id": "...", "status": "queued" }
func (h *handlers) createMigration(w http.ResponseWriter, r *http.Request) {
	userID := userIDFrom(r)
	if err := r.ParseMultipartForm(100 << 20); err != nil { // 100 MB
		http.Error(w, "parse multipart: "+err.Error(), http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("source")
	if err != nil {
		http.Error(w, "missing 'source' file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "reading upload: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if len(data) == 0 {
		http.Error(w, "empty upload", http.StatusBadRequest)
		return
	}
	if !looksLikeZip(data) {
		http.Error(w, "file does not look like a zip", http.StatusBadRequest)
		return
	}

	// Upload to Supabase Storage
	objPath := fmt.Sprintf("%s/inbound/%s", userID, header.Filename)
	if _, err := h.deps.Storage.UploadObject("source-zips", objPath, data, "application/zip"); err != nil {
		http.Error(w, "storage: "+err.Error(), http.StatusInternalServerError)
		return
	}

	pixelPerfect := r.FormValue("pixel_perfect") == "true"
	validate := r.FormValue("validate") == "true"
	strategy := r.FormValue("supabase_strategy")
	if strategy == "" {
		strategy = "extract"
	}

	id, err := db.Insert(r.Context(), h.deps.Pool, db.Migration{
		UserID:           userID,
		SourceZipPath:    objPath,
		PixelPerfect:     pixelPerfect,
		Validate:         validate,
		SupabaseStrategy: strategy,
	})
	if err != nil {
		http.Error(w, "db: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]any{"id": id, "status": db.StatusQueued})
}

// GET /api/migrations
func (h *handlers) listMigrations(w http.ResponseWriter, r *http.Request) {
	userID := userIDFrom(r)
	rows, err := db.ListByUser(r.Context(), h.deps.Pool, userID, 50)
	if err != nil {
		http.Error(w, "db: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

// GET /api/migrations/{id}
func (h *handlers) getMigration(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := userIDFrom(r)
	m, err := db.GetByID(r.Context(), h.deps.Pool, id, userID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, m)
}

// GET /api/migrations/{id}/download → 302 to signed Supabase Storage URL.
func (h *handlers) downloadOutput(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := userIDFrom(r)
	m, err := db.GetByID(r.Context(), h.deps.Pool, id, userID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if m.OutputZipPath == nil || *m.OutputZipPath == "" {
		http.Error(w, "output not ready", http.StatusConflict)
		return
	}
	signed, err := h.deps.Storage.SignedURL("output-zips", *m.OutputZipPath, 60*60)
	if err != nil {
		http.Error(w, "signed url: "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, signed, http.StatusFound)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// looksLikeZip checks the local file signature ("PK\x03\x04").
func looksLikeZip(b []byte) bool {
	return len(b) >= 4 && b[0] == 'P' && b[1] == 'K' && b[2] == 0x03 && b[3] == 0x04
}
