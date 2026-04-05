package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"um6p_fit_backend/database"
	"um6p_fit_backend/models"
)

func CreateWorkout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var workout models.Workout
	err := json.NewDecoder(r.Body).Decode(&workout)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if workout.Date.IsZero() {
		workout.Date = time.Now()
	}

	result := database.DB.Create(&workout)
	if result.Error != nil {
		http.Error(w, "Failed to save workout to database", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(workout)
}

func GetUserWorkouts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.PathValue("id")
	userID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var workouts []models.Workout
	query := database.DB.Preload("Exercises").Preload("Exercises.Exercise").Preload("Exercises.Sets").Where("user_id = ?", userID)

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
			query = query.Where("date >= ?", startDate)
		}
	}

	query = query.Order("date DESC")
	result := query.Find(&workouts)

	if result.Error != nil {
		http.Error(w, "Failed to fetch workouts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workouts)
}

func DeleteWorkout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.PathValue("id")
	workoutID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	result := database.DB.Select("Exercises").Delete(&models.Workout{ID: uint(workoutID)})

	if result.Error != nil {
		http.Error(w, "Failed to fetch workouts", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func UpdateWorkout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.PathValue("id")
	workoutID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var updateData models.Workout
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var existingWorkout models.Workout
	if result := database.DB.First(&existingWorkout, workoutID); result.Error != nil {
		http.Error(w, "Failed to fetch workouts", http.StatusInternalServerError)
		return
	}

	database.DB.Model(&existingWorkout).Updates(updateData)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingWorkout)
}

func calculateVolume(sets []models.Set) float64 {
	var totalVolume float64
	for _, s := range sets {
		totalVolume += float64(s.Reps) * s.Weight
	}
	return totalVolume
}
