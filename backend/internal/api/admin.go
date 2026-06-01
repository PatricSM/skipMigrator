package api

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// GET /api/me — returns the authenticated user's identity + role flag.
// Used by the frontend to decide whether to render the "Admin" link.
func (h *handlers) me(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"user_id":         userIDFrom(r),
		"email":           emailFrom(r),
		"is_super_admin":  isSuperAdminFrom(r),
	})
}

type adminUser struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	CreatedAt  string `json:"created_at"`
	LastSignIn string `json:"last_sign_in_at,omitempty"`
}

// GET /api/admin/users — proxies Supabase /auth/v1/admin/users.
func (h *handlers) adminListUsers(w http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf("%s/auth/v1/admin/users?per_page=100", h.deps.SupabaseURL)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	h.applyAdminHeaders(req)

	resp, err := h.adminClient().Do(req)
	if err != nil {
		http.Error(w, "upstream: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		http.Error(w, "supabase: "+string(body), resp.StatusCode)
		return
	}

	// Supabase returns { users: [ {id, email, created_at, last_sign_in_at, ...} ] }
	var raw struct {
		Users []adminUser `json:"users"`
	}
	_ = json.Unmarshal(body, &raw)
	writeJSON(w, http.StatusOK, raw.Users)
}

type createUserReq struct {
	Email    string `json:"email"`
	Password string `json:"password,omitempty"` // optional: auto-generated if empty
}

// POST /api/admin/users { email, password? } — creates a Supabase user,
// email already confirmed. Returns the generated password if not provided.
func (h *handlers) adminCreateUser(w http.ResponseWriter, r *http.Request) {
	var body createUserReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}
	if body.Email == "" {
		http.Error(w, "email required", http.StatusBadRequest)
		return
	}
	if body.Password == "" {
		body.Password = generateTempPassword(16)
	}

	payload, _ := json.Marshal(map[string]any{
		"email":         body.Email,
		"password":      body.Password,
		"email_confirm": true,
	})
	url := fmt.Sprintf("%s/auth/v1/admin/users", h.deps.SupabaseURL)
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	h.applyAdminHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.adminClient().Do(req)
	if err != nil {
		http.Error(w, "upstream: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		http.Error(w, "supabase: "+string(respBody), resp.StatusCode)
		return
	}
	var created adminUser
	_ = json.Unmarshal(respBody, &created)

	writeJSON(w, http.StatusCreated, map[string]any{
		"user":              created,
		"generated_password": body.Password, // surfaced so admin can share manually
	})
}

// DELETE /api/admin/users/{id}
func (h *handlers) adminDeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}
	url := fmt.Sprintf("%s/auth/v1/admin/users/%s", h.deps.SupabaseURL, id)
	req, _ := http.NewRequest(http.MethodDelete, url, nil)
	h.applyAdminHeaders(req)
	resp, err := h.adminClient().Do(req)
	if err != nil {
		http.Error(w, "upstream: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		http.Error(w, "supabase: "+string(b), resp.StatusCode)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *handlers) applyAdminHeaders(req *http.Request) {
	req.Header.Set("apikey", h.deps.SupabaseServiceKey)
	req.Header.Set("Authorization", "Bearer "+h.deps.SupabaseServiceKey)
}

func (h *handlers) adminClient() *http.Client {
	return &http.Client{Timeout: 0}
}

// generateTempPassword returns a URL-safe base64 string of length ~n.
// Used when admin omits the password field.
func generateTempPassword(n int) string {
	buf := make([]byte, n)
	_, _ = rand.Read(buf)
	return base64.RawURLEncoding.EncodeToString(buf)[:n]
}
