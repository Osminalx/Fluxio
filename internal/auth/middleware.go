package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/Osminalx/fluxio/internal/services"
	"github.com/Osminalx/fluxio/pkg/utils/logger"
)

type contextKey string
const UserContextKey contextKey = "user"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			logger.Warn("🚫 Intento de acceso sin token de autorización desde %s", r.RemoteAddr)
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Check if it's a Bearer token
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			logger.Warn("🚫 Formato de token inválido desde %s", r.RemoteAddr)
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}

		tokenString := tokenParts[1]

		// Validate token
		token, err := services.ValidateToken(tokenString)
		if err != nil {
			logger.Warn("🚫 Token inválido desde %s: %v", r.RemoteAddr, err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Extract claims from token
		claims, ok := token.Claims.(*services.Claims)
		if !ok {
			logger.Warn("🚫 Claims inválidos desde %s", r.RemoteAddr)
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		// Log successful authentication
		logger.Auth("ACCESS", claims.UserID, true, "Route: "+r.URL.Path)

		// Store user claims in request context
		ctx := r.Context()
		ctx = context.WithValue(ctx, "userClaims", claims)
		r = r.WithContext(ctx)

		// Call next handler
		next.ServeHTTP(w, r)
	})
}