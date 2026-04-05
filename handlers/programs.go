package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"um6p_fit_backend/database"
	"um6p_fit_backend/models"
)

func CreateWorkoutProgram(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var program models.WorkoutProgram
	if err := json.NewDecoder(r.Body).Decode(&program); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if result := database.DB.Create(&program); result.Error != nil {
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
	var userID int
	if userIDStr != "" {
		userID, _ = strconv.Atoi(userIDStr)
	}

	var programs []models.WorkoutProgram
	query := database.DB.Preload("Days")

	if userID > 0 {
		query = query.Where("created_by = 0 OR created_by = ?", userID)
	} else {
		query = query.Where("created_by = 0")
	}

	result := query.Find(&programs)

	if result.Error != nil {
		http.Error(w, "Failed to fetch workout programs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(programs)
}

func UpdateWorkoutProgram(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	programID := r.PathValue("id")
	userIDStr := r.URL.Query().Get("userId")
	if userIDStr == "" {
		http.Error(w, "Unauthorized: userId query parameter is required", http.StatusUnauthorized)
		return
	}
	userID, _ := strconv.Atoi(userIDStr)

	var existingProgram models.WorkoutProgram
	if result := database.DB.First(&existingProgram, programID); result.Error != nil {
		http.Error(w, "Program not found", http.StatusNotFound)
		return
	}

	if existingProgram.CreatedBy != uint(userID) {
		http.Error(w, "Forbidden: You do not have permission to edit this program", http.StatusForbidden)
		return
	}

	var input models.WorkoutProgram
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	existingProgram.Name = input.Name
	existingProgram.Description = input.Description
	database.DB.Save(&existingProgram)

	database.DB.Where("program_id = ?", existingProgram.ID).Delete(&models.ProgramDay{})
	for _, day := range input.Days {
		day.ProgramID = existingProgram.ID
		database.DB.Create(&day)
	}

	database.DB.Preload("Days").First(&existingProgram, existingProgram.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingProgram)
}

func DeleteWorkoutProgram(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	programID := r.PathValue("id")
	userIDStr := r.URL.Query().Get("userId")
	if userIDStr == "" {
		http.Error(w, "Unauthorized: userId query parameter is required", http.StatusUnauthorized)
		return
	}
	userID, _ := strconv.Atoi(userIDStr)

	var program models.WorkoutProgram
	if result := database.DB.First(&program, programID); result.Error != nil {
		http.Error(w, "Program not found", http.StatusNotFound)
		return
	}

	if program.CreatedBy != uint(userID) || program.CreatedBy == 0 {
		http.Error(w, "Forbidden: You can only delete your own custom programs", http.StatusForbidden)
		return
	}

	database.DB.Where("program_id = ?", program.ID).Delete(&models.ProgramDay{})
	database.DB.Where("program_id = ?", program.ID).Delete(&models.UserProgramProgress{})
	database.DB.Delete(&program)

	w.WriteHeader(http.StatusNoContent)
}

func StartProgram(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	programIDStr := r.PathValue("id")
	programID, _ := strconv.Atoi(programIDStr)

	var input struct {
		UserID uint `json:"userId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	database.DB.Model(&models.UserProgramProgress{}).
		Where("user_id = ?", input.UserID).
		Update("is_active", false)

	progress := models.UserProgramProgress{
		UserID:     input.UserID,
		ProgramID:  uint(programID),
		CurrentDay: 1,
		IsActive:   true,
	}

	if result := database.DB.Create(&progress); result.Error != nil {
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

	var input struct {
		UserID uint `json:"userId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var progress models.UserProgramProgress
	if result := database.DB.Where("user_id = ? AND is_active = true", input.UserID).Preload("Program.Days").First(&progress); result.Error != nil {
		http.Error(w, "No active program found for user", http.StatusNotFound)
		return
	}

	maxDays := len(progress.Program.Days)

	if progress.CurrentDay >= maxDays {
		progress.IsActive = false
		database.DB.Save(&progress)

		awardPoints(progress.UserID, 200, "Full Program Completed")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Program completed! Congratulations!"})
		return
	}

	progress.CurrentDay++
	database.DB.Save(&progress)

	awardPoints(progress.UserID, 20, "Program Day Completed")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(progress)
}

func GetProgramHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.PathValue("id")
	userID, _ := strconv.Atoi(userIDStr)

	var history []models.UserProgramProgress
	result := database.DB.Where("user_id = ?", userID).Preload("Program").Find(&history)

	if result.Error != nil {
		http.Error(w, "Failed to fetch program history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}
