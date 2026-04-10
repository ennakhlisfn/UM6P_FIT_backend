package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"um6p_fit_backend/database"
	"um6p_fit_backend/models"
)

// UpdateUserWeight saves a new weight float to the user profile and generates an atomic WeightLog entry
func UpdateUserWeight(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.PathValue("id")
	userID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID formatting", http.StatusBadRequest)
		return
	}

	var input struct {
		Weight float64 `json:"weight"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON request body", http.StatusBadRequest)
		return
	}

	// Override standard generic table User entry
	if result := database.DB.Model(&models.User{}).Where("id = ?", userID).Update("weight", input.Weight); result.Error != nil {
		http.Error(w, "Failed to inject weight overwrite", http.StatusInternalServerError)
		return
	}

	// Instantiate the new standalone history row 
	log := models.WeightLog{
		UserID:   uint(userID),
		Weight:   input.Weight,
		LoggedAt: time.Now(),
	}
	database.DB.Create(&log)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"message": "User weight has been correctly updated", "log": log})
}

// GetUserNotifications sweeps the notifications namespace querying all distinct alerts
func GetUserNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, _ := strconv.Atoi(r.PathValue("id"))

	var notifications []models.Notification
	if result := database.DB.Where("user_id = ?", userID).Order("created_at DESC").Limit(50).Find(&notifications); result.Error != nil {
		http.Error(w, "Failed communicating with notification backend", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifications)
}

// MarkNotificationRead converts an unread row into a boolean active flag read 
func MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	notifID := r.PathValue("id")
	
	database.DB.Model(&models.Notification{}).Where("id = ?", notifID).Update("is_read", true)
	w.WriteHeader(http.StatusOK)
}
