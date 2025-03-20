package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fragmentMetadata struct {
	Id           string
	OwnerId      string
	Created      string
	Updated      string
	FragmentType string
	Size         int
	FragmnetName string
}

type PostFragmentResponse struct {
	Location string
	message  string
	metadata fragmentMetadata
	status   string
}

func TestHealthCheck(t *testing.T) {
	setup := PreTestSetup()
	defer setup()

	// Set up and make request
	r := getRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	// Assert correct response
	assert.Equal(t, 200, w.Code)
}

func TestCacheControlHeader(t *testing.T) {
	setup := PreTestSetup()
	defer setup()

	// Set up and make request
	r := getRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	// Assert correct response
	assert.Equal(t, "no-cache", w.Result().Header.Get("Cache-Control"))
}

func TestOkInResponse(t *testing.T) {
	setup := PreTestSetup()
	defer setup()

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
	setup := PreTestSetup()
	defer setup()

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
	setup := PreTestSetup()
	defer setup()

	// Set up and make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/fragments", nil)

	r := getRouter()
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, 401, w.Result().StatusCode)
}

func TestIncorrectLoginCredentials(t *testing.T) {
	setup := PreTestSetup()
	defer setup()

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
	setup := PreTestSetup()
	defer setup()

	// Set up and make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/fragments", nil)
	req.SetBasicAuth("user1@email.com", "password1")
	r := getRouter()
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, w.Result().StatusCode, 200)
}

func TestPostFragment(t *testing.T) {
	setup := PreTestSetup()
	defer setup()

	// Set up and make request
	w := httptest.NewRecorder()

	fileData := []byte("Some test data in the file!")

	req, _ := http.NewRequest("POST", "/v1/fragments", bytes.NewReader(fileData))
	req.Header.Add("Content-Type", "text/plain")
	req.SetBasicAuth("user1@email.com", "password1")

	r := getRouter()
	r.ServeHTTP(w, req)

	fmt.Println(GetBody(w.Body.Bytes()))

	// Assert
	assert.Equal(t, 200, w.Result().StatusCode)
}

func TestGetFragment(t *testing.T) {
	setup := PreTestSetup()
	defer setup()

	// Set up and make request
	w := httptest.NewRecorder()
	fileData := []byte("Some test data in the file!")

	req, _ := http.NewRequest("POST", "/v1/fragments", bytes.NewReader(fileData))
	req.Header.Add("Content-Type", "text/plain")
	req.SetBasicAuth("user1@email.com", "password1")

	r := getRouter()
	r.ServeHTTP(w, req)
	var res PostFragmentResponse
	json.Unmarshal(w.Body.Bytes(), &res)
	fmt.Println(res)

	w = httptest.NewRecorder()
	fmt.Println("Location: ", res.Location)
	getReq, _ := http.NewRequest("GET", res.Location, nil)
	getReq.SetBasicAuth("user1@email.com", "password1")
	r.ServeHTTP(w, getReq)

	size, _ := strconv.Atoi(w.Result().Header.Get("Content-Length"))
	fmt.Println(size)
	retrievedFileBuffer := make([]byte, size)
	w.Body.Read(retrievedFileBuffer)
	fmt.Println(string(retrievedFileBuffer))

	assert.Equal(t, fileData, retrievedFileBuffer)
}
