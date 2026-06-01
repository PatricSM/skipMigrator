package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type ctxKey string

const ctxUserID ctxKey = "userID"

// authMiddleware verifies a Supabase JWT (HS256 with the project's JWT secret,
// or RS256 via JWKS). For simplicity we accept Bearer tokens and trust the `sub`
// claim — in production, verify the signature against your Supabase JWT secret.
func authMiddleware(_ Deps) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				http.Error(w, "missing bearer token", http.StatusUnauthorized)
				return
			}
			tokenStr := strings.TrimPrefix(auth, "Bearer ")

			// TODO: switch to keyfunc with JWKS for RS256 or shared HS256 secret.
			// For MVP we parse claims without signature verification and extract `sub`.
			parser := jwt.NewParser(jwt.WithoutClaimsValidation())
			claims := jwt.MapClaims{}
			_, _, err := parser.ParseUnverified(tokenStr, &claims)
			if err != nil {
				http.Error(w, "invalid jwt", http.StatusUnauthorized)
				return
			}
			sub, _ := claims["sub"].(string)
			if sub == "" {
				http.Error(w, "missing sub claim", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), ctxUserID, sub)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func userIDFrom(r *http.Request) string {
	v, _ := r.Context().Value(ctxUserID).(string)
	return v
}
