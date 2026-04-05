package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"um6p_fit_backend/database"
	"um6p_fit_backend/models"
)

func GetExercises(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var exercises []models.Exercise
	result := database.DB.Limit(5).Find(&exercises)
	if result.Error != nil {
		http.Error(w, "Failed to fetch exercises", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(exercises)
}

func GetExerciseProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	exerciseID := r.PathValue("exId")

	var workouts []models.Workout
	query := database.DB.Preload("Exercises", "exercise_id = ?", exerciseID).
		Preload("Exercises.Sets").
		Joins("JOIN workout_exercises ON workout_exercises.workout_id = workouts.id").
		Where("workouts.user_id = ? AND workout_exercises.exercise_id = ?", userID, exerciseID)

	period := r.URL.Query().Get("period")
	if period != "" {
		now := time.Now()
		var startDate time.Time
		switch period {
		case "week":
			startDate = now.AddDate(0, 0, -7)
		case "month":
			startDate = now.AddDate(0, -1, 0)
		case "year":
			startDate = now.AddDate(-1, 0, 0)
		}
		if !startDate.IsZero() {
			query = query.Where("workouts.date >= ?", startDate)
		}
	}

	result := query.Order("workouts.date ASC").Find(&workouts)

	if result.Error != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	type DataPoint struct {
		Date   string  `json:"date"`
		Volume float64 `json:"volume"`
	}

	var progressChart []DataPoint

	for _, workout := range workouts {
		if len(workout.Exercises) > 0 {
			ex := workout.Exercises[0]
			vol := calculateVolume(ex.Sets)

			progressChart = append(progressChart, DataPoint{
				Date:   workout.Date.Format("2006-01-02"),
				Volume: vol,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(progressChart)
}
