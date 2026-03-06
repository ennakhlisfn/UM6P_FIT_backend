package main

import (
	"encoding/json"
	"net/http"
	"time"
	"strconv"
)

func GetExercises(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var exercises []Exercise
	result := db.Limit(50).Find(&exercises)
	if result.Error != nil {
		http.Error(w, "Failed to fetch exercises", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(exercises)
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}

	result := db.Create(&user)
	if result.Error != nil {
		http.Error(w,"Failed to create user (Email might already exist)", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}
func CreateWorkout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var workout Workout
	err := json.NewDecoder(r.Body).Decode(&workout)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if workout.Date.IsZero() {
		workout.Date = time.Now()
	}

	result := db.Create(&workout)
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

	var workouts []Workout

	result := db.Preload("Exercises").Preload("Exercises.Exercise").Where("user_id = ?", userID).Find(&workouts)

	if result.Error != nil {
		http.Error(w, "Failed to fetch workouts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workouts)
}
