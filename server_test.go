package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
	// Set up and make request
	r := getRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	// Assert correct response
	assert.Equal(t, 200, w.Code)
}

func TestCacheControlHeader(t *testing.T) {
	// Set up and make request
	r := getRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	// Assert correct response
	assert.Equal(t, "no-cache", w.Result().Header.Get("Cache-Control"))
}

func TestOkInResponse(t *testing.T) {
	// Set up and make request
	r := getRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	// Process response to get body
	resp := w.Result()
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Error("Failed to read the response body")
	}
	jsonMap := GetBody(bodyBytes)

	// Assert correct response
	assert.Equal(t, "ok", jsonMap["status"])
}

func TestBodyInformation(t *testing.T) {
	// Set up and make request
	r := getRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	// Process response to get body
	resp := w.Result()
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Error("Failed to read the response body")
	}
	jsonMap := GetBody(bodyBytes)

	// Assert correct response
	assert.Equal(t, "https://github.com/Jashanpreet2/fragments", jsonMap["githuburl"])
	assert.Equal(t, "Jashanpreet Singh", jsonMap["author"])
	assert.Equal(t, "1", jsonMap["version"])
}

func TestUnauthenticatedRequest(t *testing.T) {
	// Set up and make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/fragments", nil)
	r := getRouter()
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, 401, w.Result().StatusCode)
}

func TestIncorrectLoginCredentials(t *testing.T) {
	// Set up and make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/fragments", nil)
	req.SetBasicAuth("invalid@email.com", "incorrect_password")
	r := getRouter()
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, w.Result().StatusCode, 401)
}

func TestAuthenticatedUser(t *testing.T) {
	// Set up and make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/fragments", nil)
	req.SetBasicAuth("user1@email.com", "password1")
	r := getRouter()
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, w.Result().StatusCode, 200)
}
