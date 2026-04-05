package handlers

import (
	"encoding/json"
	"net/http"

	"um6p_fit_backend/database"
	"um6p_fit_backend/models"
)

func GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var leaderboard []models.LeaderboardEntry

	err := database.DB.Table("users").
		Select("users.id, users.name, COALESCE(SUM(sets.reps * sets.weight), 0) as total_volume").
		Joins("LEFT JOIN workouts ON workouts.user_id = users.id").
		Joins("LEFT JOIN workout_exercises ON workout_exercises.workout_id = workouts.id").
		Joins("LEFT JOIN sets ON sets.workout_exercise_id = workout_exercises.id").
		Group("users.id, users.name").
		Order("total_volume DESC").
		Limit(10).
		Scan(&leaderboard).Error

	if err != nil {
		http.Error(w, "Failed to generate leaderboard", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leaderboard)
}
