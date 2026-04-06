package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Since true DB testing requires mocking out GORM completely,
// we will verify that the routing and struct responses shape out correctly.
func TestExpectedShape_AdminGetCommunityRank(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/admin/stats/community-rank", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(AdminGetCommunityRank)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode JSON response: %v", err)
	}

	if _, exists := response["communityScore"]; !exists {
		t.Errorf("Expected key 'communityScore' not present in JSON output")
	}

	if _, exists := response["globalRank"]; !exists {
		t.Errorf("Expected key 'globalRank' not present in JSON output")
	}

	if _, exists := response["syncTime"]; !exists {
		t.Errorf("Expected key 'syncTime' not present in JSON output")
	}
}

func TestExpectedShape_AdminGetLiveClasses(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/admin/live-classes/active", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(AdminGetLiveClasses)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK")
	}

	var response map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&response)

	if _, exists := response["activeNow"]; !exists {
		t.Errorf("Expected key 'activeNow' not present")
	}
}
