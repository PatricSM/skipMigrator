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
func (h *handlers) me(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"user_id":        userIDFrom(r),
		"email":          emailFrom(r),
		"is_super_admin": isSuperAdminFrom(r),
	})
}

type adminUser struct {
	ID          string         `json:"id"`
	Email       string         `json:"email"`
	CreatedAt   string         `json:"created_at"`
	LastSignIn  string         `json:"last_sign_in_at,omitempty"`
	AppMetadata map[string]any `json:"app_metadata,omitempty"`
}

// listResponseUser is the trimmed payload we return to the frontend, with the
// role bubbled up so the UI can show a badge without parsing app_metadata.
type listResponseUser struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	CreatedAt  string `json:"created_at"`
	LastSignIn string `json:"last_sign_in_at,omitempty"`
	Role       string `json:"role"` // "super_admin" | "user"
}

func roleFromAppMeta(m map[string]any) string {
	if r, ok := m["role"].(string); ok && r == "super_admin" {
		return "super_admin"
	}
	return "user"
}

// GET /api/admin/users — proxies Supabase /auth/v1/admin/users.
// Augments each user with role: 'super_admin' if either app_metadata.role is set
// OR the email is in SUPER_ADMIN_EMAILS env.
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

	var raw struct {
		Users []adminUser `json:"users"`
	}
	_ = json.Unmarshal(body, &raw)

	envAdmins := map[string]bool{}
	for _, e := range h.deps.SuperAdminEmails {
		envAdmins[e] = true
	}
	out := make([]listResponseUser, 0, len(raw.Users))
	for _, u := range raw.Users {
		role := roleFromAppMeta(u.AppMetadata)
		if envAdmins[u.Email] {
			role = "super_admin"
		}
		out = append(out, listResponseUser{
			ID:         u.ID,
			Email:      u.Email,
			CreatedAt:  u.CreatedAt,
			LastSignIn: u.LastSignIn,
			Role:       role,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

type createUserReq struct {
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
	Role     string `json:"role,omitempty"` // "user" (default) or "super_admin"
}

// POST /api/admin/users { email, password?, role? } — creates a Supabase user
// already confirmed, with the requested role stored in app_metadata.
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
	if body.Role != "super_admin" {
		body.Role = "user"
	}

	payload, _ := json.Marshal(map[string]any{
		"email":         body.Email,
		"password":      body.Password,
		"email_confirm": true,
		"app_metadata":  map[string]any{"role": body.Role},
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
		"user": listResponseUser{
			ID:         created.ID,
			Email:      created.Email,
			CreatedAt:  created.CreatedAt,
			LastSignIn: created.LastSignIn,
			Role:       body.Role,
		},
		"generated_password": body.Password,
	})
}

type updateRoleReq struct {
	Role string `json:"role"` // "user" or "super_admin"
}

// PATCH /api/admin/users/{id}/role { role } — updates app_metadata.role.
// The user must sign out + sign back in for the new role to appear in their JWT.
func (h *handlers) adminUpdateRole(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}
	var body updateRoleReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if body.Role != "super_admin" && body.Role != "user" {
		http.Error(w, "role must be 'user' or 'super_admin'", http.StatusBadRequest)
		return
	}

	payload, _ := json.Marshal(map[string]any{
		"app_metadata": map[string]any{"role": body.Role},
	})
	url := fmt.Sprintf("%s/auth/v1/admin/users/%s", h.deps.SupabaseURL, id)
	req, _ := http.NewRequest(http.MethodPut, url, bytes.NewReader(payload))
	h.applyAdminHeaders(req)
	req.Header.Set("Content-Type", "application/json")

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
	writeJSON(w, http.StatusOK, map[string]any{"id": id, "role": body.Role})
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
func generateTempPassword(n int) string {
	buf := make([]byte, n)
	_, _ = rand.Read(buf)
	return base64.RawURLEncoding.EncodeToString(buf)[:n]
}
