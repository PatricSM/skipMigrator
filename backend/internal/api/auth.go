package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type ctxKey string

const (
	ctxUserID  ctxKey = "userID"
	ctxEmail   ctxKey = "email"
	ctxIsAdmin ctxKey = "isSuperAdmin"
)

// authMiddleware parses the bearer JWT and stuffs `sub`, `email`, and
// `isSuperAdmin` (email ∈ SUPER_ADMIN_EMAILS) into the request context.
//
// For MVP we don't verify the signature: we trust Supabase's reverse proxy
// and rely on RLS at the data layer. Production hardening would verify
// against the project's JWT secret (HS256) or JWKS (RS256).
func authMiddleware(deps Deps) func(http.Handler) http.Handler {
	adminSet := map[string]bool{}
	for _, e := range deps.SuperAdminEmails {
		adminSet[strings.ToLower(strings.TrimSpace(e))] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				http.Error(w, "missing bearer token", http.StatusUnauthorized)
				return
			}
			tokenStr := strings.TrimPrefix(auth, "Bearer ")

			parser := jwt.NewParser(jwt.WithoutClaimsValidation())
			claims := jwt.MapClaims{}
			if _, _, err := parser.ParseUnverified(tokenStr, &claims); err != nil {
				http.Error(w, "invalid jwt", http.StatusUnauthorized)
				return
			}
			sub, _ := claims["sub"].(string)
			email, _ := claims["email"].(string)
			if sub == "" {
				http.Error(w, "missing sub claim", http.StatusUnauthorized)
				return
			}

			// Super admin can come from either:
			//   (a) email is in SUPER_ADMIN_EMAILS env (bootstrap)
			//   (b) JWT carries app_metadata.role = "super_admin" (set via admin API)
			isAdmin := email != "" && adminSet[strings.ToLower(email)]
			if !isAdmin {
				if appMeta, ok := claims["app_metadata"].(map[string]any); ok {
					if role, _ := appMeta["role"].(string); role == "super_admin" {
						isAdmin = true
					}
				}
			}

			ctx := context.WithValue(r.Context(), ctxUserID, sub)
			ctx = context.WithValue(ctx, ctxEmail, email)
			ctx = context.WithValue(ctx, ctxIsAdmin, isAdmin)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// requireSuperAdmin gates routes that only super admins can access.
// Must run after authMiddleware so the context has isSuperAdmin set.
func requireSuperAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if v, _ := r.Context().Value(ctxIsAdmin).(bool); !v {
			http.Error(w, "forbidden: super admin only", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func userIDFrom(r *http.Request) string {
	v, _ := r.Context().Value(ctxUserID).(string)
	return v
}

func emailFrom(r *http.Request) string {
	v, _ := r.Context().Value(ctxEmail).(string)
	return v
}

func isSuperAdminFrom(r *http.Request) bool {
	v, _ := r.Context().Value(ctxIsAdmin).(bool)
	return v
}
