package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"um6p_fit_backend/middleware"
)

func setupTestRouter() *http.ServeMux {
	mux := http.NewServeMux()

	// Seed public routes
	mux.HandleFunc("POST /api/users", CreateUser)
	mux.HandleFunc("POST /api/login", LoginUser)
	mux.HandleFunc("GET /api/leaderboard", GetLeaderboard)
	mux.HandleFunc("GET /api/exercises", GetExercises)

	// Seed secure routes wrapped by actual Auth Middleware
	mux.HandleFunc("POST /api/workouts", middleware.AuthMiddleware(CreateWorkout))
	mux.HandleFunc("PUT /api/users/{id}/weight", middleware.AuthMiddleware(UpdateUserWeight))
	mux.HandleFunc("GET /api/users/{id}/notifications", middleware.AuthMiddleware(GetUserNotifications))
	mux.HandleFunc("PUT /api/notifications/{id}/read", middleware.AuthMiddleware(MarkNotificationRead))
	mux.HandleFunc("POST /api/programs", middleware.AuthMiddleware(CreateWorkoutProgram))
	mux.HandleFunc("GET /api/programs", GetWorkoutPrograms)

	return mux
}

// Helper: Sets up a user and returns their Auth Token and dynamic User ID
func setupTestUser(mux *http.ServeMux, email string) (string, int) {
	// Create
	userPayload := fmt.Sprintf(`{"name": "Tester", "email": "%s", "password": "pass", "age": 22, "height": 180, "weight": 70}`, email)
	req1, _ := http.NewRequest("POST", "/api/users", bytes.NewBufferString(userPayload))
	req1.Header.Set("Content-Type", "application/json")
	rr1 := httptest.NewRecorder()
	mux.ServeHTTP(rr1, req1)

	// Login
	loginPayload := fmt.Sprintf(`{"email": "%s", "password": "pass"}`, email)
	req2, _ := http.NewRequest("POST", "/api/login", bytes.NewBufferString(loginPayload))
	req2.Header.Set("Content-Type", "application/json")
	rr2 := httptest.NewRecorder()
	mux.ServeHTTP(rr2, req2)

	var loginData map[string]interface{}
	json.NewDecoder(rr2.Body).Decode(&loginData)

	token := loginData["token"].(string)
	userData := loginData["user"].(map[string]interface{})
	userID := int(userData["id"].(float64))

	return token, userID
}

func TestCoreIntegrationFlow(t *testing.T) {
	SetupTestDB()
	mux := setupTestRouter()

	token, userID := setupTestUser(mux, "core@test.com")

	// -----------------------------------------
	// 1. Post a Secure Workout (Protected)
	// -----------------------------------------
	workoutPayload := map[string]interface{}{
		"userId": userID,
		"name": "Integration Sprint",
		"exercises": []map[string]interface{}{
			{
				"exerciseId": "1_2_Squat",
				"sets": []map[string]interface{}{
					{"reps": 10, "weight": 50.0},
				},
			},
		},
	}
	workoutBytes, _ := json.Marshal(workoutPayload)
	req3, _ := http.NewRequest("POST", "/api/workouts", bytes.NewBuffer(workoutBytes))
	req3.Header.Set("Content-Type", "application/json")
	req3.Header.Set("Authorization", "Bearer "+token)
	
	rr3 := httptest.NewRecorder()
	mux.ServeHTTP(rr3, req3)

	if rr3.Code != http.StatusCreated && rr3.Code != http.StatusOK {
		t.Fatalf("Failed to post secure workout via auth token. Status: %v", rr3.Code)
	}

	// -----------------------------------------
	// 2. Verify Leaderboard Points Generation
	// -----------------------------------------
	req4, _ := http.NewRequest("GET", "/api/leaderboard", nil)
	rr4 := httptest.NewRecorder()
	mux.ServeHTTP(rr4, req4)

	var leaderboard []map[string]interface{}
	json.NewDecoder(rr4.Body).Decode(&leaderboard)

	if len(leaderboard) != 1 {
		t.Fatalf("Expected 1 person on global test leaderboard, got %d", len(leaderboard))
	}

	if _, ok := leaderboard[0]["totalPoints"]; !ok {
		t.Errorf("Leaderboard aggregation failed! Expected totalPoints key, got: %v", leaderboard[0])
	}
}

func TestWeightAndNotificationFlow(t *testing.T) {
	SetupTestDB()
	mux := setupTestRouter()
	token, userID := setupTestUser(mux, "notif@test.com")

	// 1. Update Weight
	weightPayload := `{"weight": 78.5}`
	wReq, _ := http.NewRequest("PUT", fmt.Sprintf("/api/users/%d/weight", userID), bytes.NewBufferString(weightPayload))
	wReq.Header.Set("Content-Type", "application/json")
	wReq.Header.Set("Authorization", "Bearer "+token)

	wRR := httptest.NewRecorder()
	mux.ServeHTTP(wRR, wReq)

	if wRR.Code != http.StatusOK {
		t.Fatalf("Failed to execute Update Weight API. Status: %v", wRR.Code)
	}

	// 2. Fetch Notifications 
	nReq, _ := http.NewRequest("GET", fmt.Sprintf("/api/users/%d/notifications", userID), nil)
	nReq.Header.Set("Authorization", "Bearer "+token)

	nRR := httptest.NewRecorder()
	mux.ServeHTTP(nRR, nReq)

	if nRR.Code != http.StatusOK {
		t.Fatalf("Failed to get Notifications. Status: %v", nRR.Code)
	}
}

func TestProgramsAPI(t *testing.T) {
	SetupTestDB()
	mux := setupTestRouter()
	token, userID := setupTestUser(mux, "programs@test.com")

	// 1. Create a Custom Program
	programPayload := map[string]interface{}{
		"name": "Testing Program",
		"description": "Just making sure the API natively accepts structural program mapping",
		"createdBy": userID,
		"days": []map[string]interface{}{
			{"dayNumber": 1, "workoutTemplateId": 1},
		},
	}
	bytesPayload, _ := json.Marshal(programPayload)
	req1, _ := http.NewRequest("POST", "/api/programs", bytes.NewBuffer(bytesPayload))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Authorization", "Bearer "+token)

	rr1 := httptest.NewRecorder()
	mux.ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusCreated && rr1.Code != http.StatusOK {
		t.Fatalf("Failed to dynamically allocate a Program! Status: %v", rr1.Code)
	}

	// 2. Fetch global Programs
	req2, _ := http.NewRequest("GET", "/api/programs", nil)
	rr2 := httptest.NewRecorder()
	mux.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Fatalf("Failed to fetch programs pool. Status: %v", rr2.Code)
	}
}

func TestStaticExercisesAPI(t *testing.T) {
	SetupTestDB()
	mux := setupTestRouter()

	req, _ := http.NewRequest("GET", "/api/exercises", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	// Testing simple public endpoints
	if rr.Code != http.StatusOK {
		t.Fatalf("Failed to retrieve general static DB exercises! Status: %v", rr.Code)
	}
}
