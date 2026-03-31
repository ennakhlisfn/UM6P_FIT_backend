package main

import (
	"encoding/json"
	"net/http"
	"time"
	"strconv"
    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"
)

// TODO: move it to .env
var jwtKey = []byte("my_super_secret_um6p_fit_key")

func GetExercises(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var exercises []Exercise
	result := db.Limit(5).Find(&exercises)
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

	var input struct {
		Name     string  `json:"name"`
		Email    string  `json:"email"`
		Password string  `json:"password"`
		Age      int     `json:"age"`
		Height   float64 `json:"height"`
		Weight   float64 `json:"weight"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), 14)
	if err != nil {
		http.Error(w, "Failed to encrypt password", http.StatusInternalServerError)
		return
	}

	newUser := User{
		Name:     input.Name,
		Email:    input.Email,
		Password: string(hashedPassword),
		Age:      input.Age,
		Height:   input.Height,
		Weight:   input.Weight,
	}

	if result := db.Create(&newUser); result.Error != nil {
		http.Error(w, "Failed to create user (Email might already exist)", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newUser)
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

	result := db.Select("Exercises").Delete(&Workout{ID: uint(workoutID)})

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

	var updateData Workout
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var existingWorkout Workout
	if result := db.First(&existingWorkout, workoutID); result.Error != nil {
		http.Error(w, "Failed to fetch workouts", http.StatusInternalServerError)
		return
	}

	db.Model(&existingWorkout).Updates(updateData)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingWorkout)
}

func calculateVolume(sets []Set) float64 {
	var totalVolume float64
	for _, s := range sets {
		totalVolume += float64(s.Reps) * s.Weight
	}
	return totalVolume
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

	var workouts []Workout
	result := db.Preload("Exercises", "exercise_id = ?", exerciseID).
		Joins("JOIN workout_exercises ON workout_exercises.workout_id = workouts.id").
		Where("workouts.user_id = ? AND workout_exercises.exercise_id = ?", userID, exerciseID).
		Order("workouts.date ASC").
		Find(&workouts)

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

func GetWorkoutTemplates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.URL.Query().Get("userId")

	var templates []WorkoutTemplate

	query := db.Preload("Exercises")

	if userIDStr != "" {
		userID, err := strconv.Atoi(userIDStr)
		if err == nil {
			query = query.Where("created_by = ? OR created_by = ?", 0, userID)
		} else {
			query = query.Where("created_by = ?", 0)
		}
	} else {
		query = query.Where("created_by = ?", 0)
	}

	if result := query.Find(&templates); result.Error != nil {
		http.Error(w, "Failed to fetch templates", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templates)
}

func CreateWorkoutTemplate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var template WorkoutTemplate
	if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if result := db.Create(&template); result.Error != nil {
		http.Error(w, "Failed to create workout template", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(template)
}

func UpdateWorkoutTemplate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	templateID := r.PathValue("id")
	userIDStr := r.URL.Query().Get("userId")
	if userIDStr == "" {
		http.Error(w, "Unauthorized: userId query parameter is required", http.StatusUnauthorized)
		return
	}
	userID, _ := strconv.Atoi(userIDStr)

	var existingTemplate WorkoutTemplate
	if result := db.First(&existingTemplate, templateID); result.Error != nil {
		http.Error(w, "Template not found", http.StatusNotFound)
		return
	}

	if existingTemplate.CreatedBy != uint(userID) {
		http.Error(w, "Forbidden: You do not have permission to edit this template", http.StatusForbidden)
		return
	}

	var input WorkoutTemplate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	existingTemplate.Name = input.Name
	existingTemplate.Type = input.Type
	db.Save(&existingTemplate)

	db.Where("template_id = ?", existingTemplate.ID).Delete(&TemplateExercise{})
	for _, ex := range input.Exercises {
		ex.TemplateID = existingTemplate.ID 
		db.Create(&ex)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingTemplate)
}

func DeleteWorkoutTemplate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	templateID := r.PathValue("id")
	userIDStr := r.URL.Query().Get("userId")
	if userIDStr == "" {
		http.Error(w, "Unauthorized: userId query parameter is required", http.StatusUnauthorized)
		return
	}
	userID, _ := strconv.Atoi(userIDStr)

	var template WorkoutTemplate
	if result := db.First(&template, templateID); result.Error != nil {
		http.Error(w, "Template not found", http.StatusNotFound)
		return
	}

	if template.CreatedBy != uint(userID) || template.CreatedBy == 0 {
		http.Error(w, "Forbidden: You can only delete your own custom templates", http.StatusForbidden)
		return
	}

	db.Where("template_id = ?", template.ID).Delete(&TemplateExercise{})
	db.Delete(&template)

	w.WriteHeader(http.StatusNoContent)
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var user User
	if result := db.Where("email = ?", credentials.Email).First(&user); result.Error != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password))
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

    expirationTime := time.Now().Add(24 * time.Hour)

    claims := jwt.MapClaims{
        "userId":   user.ID,
        "exp":      expirationTime.Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

    tokenString, err := token.SignedString(jwtKey)
    if err != nil {
        http.Error(w, "Failed to generate token", http.StatusInternalServerError)
        return
    }

    response := struct {
		Token string `json:"token"`
		User  User   `json:"user"`
	}{
		Token: tokenString,
		User:  user,
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GetLeaderboard(w http.ResponseWriter, r *http.Request){
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var users []User
    if result := db.Find(&users); result.Error != nil {
        http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
        return
    }

    var leaderboard []LeaderboardEntry
    for _, user := range users {
        var workouts []Workout
        db.Where("user_id = ?", user.ID).Find(&workouts)

        var totalVolume float64

        for _, workout := range workouts {
            var exercises []WorkoutExercise
            db.Where("workout_id = ?", workout.ID).Find(&exercises)

            for _, ex := range exercises {
                for _, set := range ex.Sets {
                    totalVolume += float64(set.Reps) * set.Weight
                }
            }
        }

        //if totalVolume > 0 {
            leaderboard = append(leaderboard, LeaderboardEntry{
                ID:             user.ID,
                Name:           user.Name,
                TotalVolume:    totalVolume,
            })
        //}
    }

    w.Header().Set("Content_Type", "application/json")
    json.NewEncoder(w).Encode(leaderboard)
}

func CreateWorkoutProgram(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var program WorkoutProgram
    if err := json.NewDecoder(r.Body).Decode(&program); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    if result := db.Create(&program); result.Error != nil {
        http.Error(w, "Failed to create workout program", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(program)
}

func GetWorkoutPrograms(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    userIDStr := r.URL.Query().Get("userId")
    var programs []WorkoutProgram

    query := db.Preload("Days")

    if userIDStr != "" {
        userID, err := strconv.Atoi(userIDStr)
        if err == nil {
            query = query.Where("created_by = ? OR created_by =  ?", 0, userID)
        } else {
            query = query.Where("created_by = ?", 0)
        }
    } else {
        query = query.Where("created_by = ?", 0)
    }

    if result := query.Find(&programs); result.Error != nil {
        http.Error(w, "Failed to fetch programs", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(programs)
}

func StartProgram(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    programIDStr := r.PathValue("id")
    userIDStr := r.URL.Query().Get("userId")

    if userIDStr == "" {
        http.Error(w, "Unauthorized: userId is required", http.StatusUnauthorized)
        return
    }

    programID, _ := strconv.Atoi(programIDStr)
    userID, _ := strconv.Atoi(userIDStr)

    db.Model(&UserProgramProgress{}).Where("user_id = ?", userID).Update("is_active", false)

    progress := UserProgramProgress {
        UserID:         uint(userID),
        ProgramID:      uint(programID),
        CurrentDay:     1,
        IsActive:       true,
    }

    if result := db.Create(&progress); result.Error != nil {
        http.Error(w, "Failed to start program", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(progress)
}

func AdvanceProgramDay(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    userIDStr := r.URL.Query().Get("userId")
    if userIDStr == "" {
        http.Error(w, "Unauthorized: userId is required",  http.StatusUnauthorized)
        return
    }

    userID, _ := strconv.Atoi(userIDStr)

    var progress UserProgramProgress
    if result := db.Where("user_id = ? AND is_active = ?", userID, true).First(&progress); result.Error != nil {
        http.Error(w, "No active program found for this user", http.StatusNotFound)
        return
    }

    var totalDays int64
    db.Model(&ProgramDay{}).Where("program_id = ?", progress.ProgramID).Count(&totalDays)

    if int64(progress.CurrentDay) >= totalDays {
        progress.IsActive = false
    } else {
        progress.CurrentDay += 1
    }

    if result := db.Save(&progress); result.Error != nil {
        http.Error(w, "Failed to update progress", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(progress)
}

func GetProgramHistory(w http.ResponseWriter, r *http.Request) {
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

	var programs []UserProgramProgress

	result := db.Preload("Program").Where("user_id = ? AND is_active = ?", userID, false).Find(&programs)
	if result.Error != nil {
		http.Error(w, "Failed to fetch workouts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(programs)
}
