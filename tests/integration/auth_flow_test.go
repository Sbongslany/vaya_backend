package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFullAuthenticationFlow(t *testing.T) {
	suite := SetupSuite(t)

	// --- Step 1: Register ---
	registerPayload := map[string]interface{}{
		"first_name": "John",
		"last_name":  "Doe",
		"email":      "john.doe@example.com",
		"password":   "SecurePassword123!",
	}
	body, _ := json.Marshal(registerPayload)
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.Engine.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}
	t.Log("✅ Registration successful")

	// --- Step 2: Login ---
	loginPayload := map[string]interface{}{
		"email":    "john.doe@example.com",
		"password": "SecurePassword123!",
	}
	body, _ = json.Marshal(loginPayload)
	req, _ = http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	suite.Engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var loginRes map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &loginRes)
	accessToken := loginRes["access_token"].(string)
	t.Log("✅ Login successful, received JWT")

	// --- Step 3: Get Profile (Protected Route) ---
	req, _ = http.NewRequest("GET", "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	w = httptest.NewRecorder()
	suite.Engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 for /me, got %d: %s", w.Code, w.Body.String())
	}

	var meRes map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &meRes)

	if meRes["email"] != "john.doe@example.com" {
		t.Errorf("expected email john.doe@example.com, got %v", meRes["email"])
	}
	t.Log("✅ Protected route /me accessible with valid JWT")

	// --- Step 4: Logout ---
	req, _ = http.NewRequest("POST", "/api/v1/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	w = httptest.NewRecorder()
	suite.Engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 for logout, got %d: %s", w.Code, w.Body.String())
	}
	t.Log("✅ Logout successful, session revoked")

	// --- Step 5: Verify Refresh Token is Revoked ---
	refreshToken := loginRes["refresh_token"].(string)
	refreshPayload := map[string]interface{}{
		"refresh_token": refreshToken,
	}
	body, _ = json.Marshal(refreshPayload)
	req, _ = http.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	suite.Engine.ServeHTTP(w, req)

	// Because the session was revoked, the refresh token MUST be rejected
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 for revoked refresh token, got %d", w.Code)
	}
	t.Log("✅ Refresh token correctly rejected after logout")
}
