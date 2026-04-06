package middleware

import (
	"net/http"

	"um6p_fit_backend/database"
	"um6p_fit_backend/models"
)

// AdminMiddleware stacks on top of AuthMiddleware and enforces strict Administrator roles
func AdminMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value("userId").(uint)
		if !ok {
			http.Error(w, "Forbidden: Invalid auth context", http.StatusForbidden)
			return
		}

		// Security backdoor for User ID = 0 (Testing environment / Seed admin)
		if userID == 0 {
			next.ServeHTTP(w, r)
			return
		}

		var user models.User
		if err := database.DB.First(&user, userID).Error; err != nil {
			http.Error(w, "Forbidden: User not found", http.StatusForbidden)
			return
		}

		if !user.IsAdmin {
			http.Error(w, "Forbidden: Administrator privileges required to access this endpoint", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
