package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"time"

	"um6p_fit_backend/database"
	"um6p_fit_backend/models"
)

func GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "weekly"
	}

	var leaderboard []models.LeaderboardEntry

	query := database.DB.Table("users").
		Select("users.id, users.name, COALESCE(SUM(user_points_logs.points), 0) as total_points, 0.0 as score").
		Joins("LEFT JOIN user_points_logs ON user_points_logs.user_id = users.id")

	now := time.Now()
	var startDate time.Time

	switch period {
	case "weekly":
		startDate = now.AddDate(0, 0, -7)
	case "monthly":
		startDate = now.AddDate(0, -1, 0)
	case "yearly":
		startDate = now.AddDate(-1, 0, 0)
	}

	if !startDate.IsZero() {
		query = query.Where("user_points_logs.earned_at >= ?", startDate)
	}

	if period != "alltime" {
		err := query.Group("users.id, users.name").
			Order("total_points DESC").
			Limit(10).
			Scan(&leaderboard).Error
		if err != nil {
			http.Error(w, "Failed to generate leaderboard", http.StatusInternalServerError)
			return
		}
		for i := range leaderboard {
			leaderboard[i].Score = float64(leaderboard[i].TotalPoints)
		}
	} else {
		var allUsers []models.LeaderboardEntry
		err := query.Group("users.id, users.name").
			Scan(&allUsers).Error
		if err != nil {
			http.Error(w, "Failed to fetch alltime stats", http.StatusInternalServerError)
			return
		}

		for i := range allUsers {
			var firstLog time.Time
			database.DB.Table("user_points_logs").
				Select("MIN(earned_at)").
				Where("user_id = ?", allUsers[i].ID).
				Scan(&firstLog)

			if !firstLog.IsZero() {
				days := time.Since(firstLog).Hours() / 24.0
				if days < 1 {
					days = 1
				}
				allUsers[i].Score = float64(allUsers[i].TotalPoints) / days
			} else {
				allUsers[i].Score = float64(allUsers[i].TotalPoints)
			}
		}

		sort.Slice(allUsers, func(i, j int) bool {
			return allUsers[i].Score > allUsers[j].Score // Sort DESC
		})

		if len(allUsers) > 10 {
			leaderboard = allUsers[:10]
		} else {
			leaderboard = allUsers
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leaderboard)
}
