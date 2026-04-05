package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"um6p_fit_backend/database"
	"um6p_fit_backend/models"
)

func CreateWorkoutTemplate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var template models.WorkoutTemplate
	if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if template.CreatedBy == 0 {
		template.CreatedBy = 0
	}

	if result := database.DB.Create(&template); result.Error != nil {
		http.Error(w, "Failed to create workout template", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(template)
}

func GetWorkoutTemplates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.URL.Query().Get("userId")
	if userIDStr == "" {
		http.Error(w, "Unauthorized: userId query parameter is required", http.StatusUnauthorized)
		return
	}
	userID, _ := strconv.Atoi(userIDStr)

	var templates []models.WorkoutTemplate
	result := database.DB.Preload("Exercises").Where("created_by = 0 OR created_by = ?", userID).Find(&templates)

	if result.Error != nil {
		http.Error(w, "Failed to fetch workout templates", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templates)
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

	var existingTemplate models.WorkoutTemplate
	if result := database.DB.First(&existingTemplate, templateID); result.Error != nil {
		http.Error(w, "Template not found", http.StatusNotFound)
		return
	}

	if existingTemplate.CreatedBy == 0 || existingTemplate.CreatedBy != uint(userID) {
		http.Error(w, "Forbidden: You do not have permission to edit this template", http.StatusForbidden)
		return
	}

	var input models.WorkoutTemplate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	existingTemplate.Name = input.Name
	existingTemplate.Type = input.Type
	database.DB.Save(&existingTemplate)

	database.DB.Where("template_id = ?", existingTemplate.ID).Delete(&models.TemplateExercise{})

	for _, ex := range input.Exercises {
		ex.TemplateID = existingTemplate.ID
		database.DB.Create(&ex)
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

	var template models.WorkoutTemplate
	if result := database.DB.First(&template, templateID); result.Error != nil {
		http.Error(w, "Template not found", http.StatusNotFound)
		return
	}

	if template.CreatedBy == 0 || template.CreatedBy != uint(userID) {
		http.Error(w, "Forbidden: You can only delete your own custom templates", http.StatusForbidden)
		return
	}

	database.DB.Where("template_id = ?", template.ID).Delete(&models.TemplateExercise{})
	database.DB.Delete(&template)

	w.WriteHeader(http.StatusNoContent)
}
