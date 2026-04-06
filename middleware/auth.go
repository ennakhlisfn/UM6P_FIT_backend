package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var JwtKey = []byte("my_super_secret_um6p_fit_key")

// AuthMiddleware validates JWT tokens for protected routes
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, http.ErrNotSupported
			}
			return JwtKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Unauthorized: invalid token claims", http.StatusUnauthorized)
			return
		}

		userIDFloat, ok := claims["userId"].(float64)
		if !ok {
			http.Error(w, "Unauthorized: no userId in token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "userId", uint(userIDFloat))
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
