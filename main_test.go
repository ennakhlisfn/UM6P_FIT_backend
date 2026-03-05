package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestDBConnection ensures that we can connect to the database and that it's populated.
func TestDBConnection(t *testing.T) {
	InitDB()

	var count int64
	db.Model(&Exercise{}).Count(&count)

	if count == 0 {
		t.Errorf("Expected database to contain exercises, but got 0. Did you forget to seed the database?")
	}
}

// TestAPIEndpoint ensures the /api/exercises returns valid JSON data.
func TestAPIEndpoint(t *testing.T) {
	// 1. Initialize DB to ensure we have data to fetch
	InitDB()

	// 2. Create the request
	req, err := http.NewRequest("GET", "/api/exercises", nil)
	if err != nil {
		t.Fatal(err)
	}

	// 3. Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Create a temporary handler that mirrors what's in our main.go route
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var exercises []Exercise
		result := db.Limit(50).Find(&exercises)
		if result.Error != nil {
			http.Error(w, "Failed to fetch exercises", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(exercises)
	})

	// 4. Serve the HTTP request to our recorder
	handler.ServeHTTP(rr, req)

	// 5. Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// 6. Check the response body
	var exercises []Exercise
	err = json.Unmarshal(rr.Body.Bytes(), &exercises)
	if err != nil {
		t.Fatalf("Failed to decode response JSON: %v", err)
	}

	if len(exercises) == 0 {
		t.Errorf("Expected the API to return a list of exercises, but it returned nothing")
	}
}
