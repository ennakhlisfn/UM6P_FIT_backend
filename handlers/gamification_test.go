package handlers

import (
	"testing"
	"um6p_fit_backend/database"
	"um6p_fit_backend/models"
)

func TestAwardPoints(t *testing.T) {
	// Initialize the pure Go SQLite mock database globally
	SetupTestDB()

	// Seed a dummy user for the points
	database.DB.Create(&models.User{ID: 99, Name: "Gamification Tester", Email: "game@example.com"})

	// Execute the core logic
	awardPoints(99, 150, "Completed First Workout")

	// Perform database assertion
	var log models.UserPointsLog
	database.DB.Where("user_id = ?", 99).First(&log)

	if log.Points != 150 {
		t.Errorf("Expected 150 points, got %v", log.Points)
	}
	
	if log.Reason != "Completed First Workout" {
		t.Errorf("Expected 'Completed First Workout', got %v", log.Reason)
	}
}
