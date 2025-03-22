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

type FragmentMetadata struct {
	Id           string
	OwnerId      string
	Created      string
	Updated      string
	FragmentType string
	Size         int
}

type PostFragmentResponse struct {
	Location string
	Message  string
	Metadata FragmentMetadata
	Status   string
}

type GetFragmentInfoResponse struct {
	Status   string
	Fragment FragmentMetadata
}

type GetFragmentsResponse struct {
	Status       string
	Fragment_ids []string
}

type GetFragmentsExpandedResponse struct {
	Status    string
	Fragments []FragmentMetadata
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
	req, _ := http.NewRequest("GET", "/", nil)
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

func TestGetFragments(t *testing.T) {
	setup := PreTestSetup()
	defer setup()

	r := getRouter()
	fileData := []byte("Sample data")
	mimeType := "text/plain"
	username := "user1@email.com"
	password := "password1"

	postFragmentResponse := PostFragment(r, fileData, mimeType, username, password)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/fragments", nil)
	req.SetBasicAuth(username, password)

	r.ServeHTTP(w, req)
	var response GetFragmentsResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, response.Fragment_ids[0], postFragmentResponse.Metadata.Id)
}

func TestGetFragmentsExpanded(t *testing.T) {
	setup := PreTestSetup()
	defer setup()

	r := getRouter()
	fileData := []byte("Sample data")
	mimeType := "text/plain"
	username := "user1@email.com"
	password := "password1"

	postFragmentResponse := PostFragment(r, fileData, mimeType, username, password)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/fragments?expand=1", nil)
	req.SetBasicAuth(username, password)

	r.ServeHTTP(w, req)
	var response GetFragmentsExpandedResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, response.Fragments[0], postFragmentResponse.Metadata)
}

func TestGetFragment(t *testing.T) {
	setup := PreTestSetup()
	defer setup()

	r := getRouter()
	fileData := []byte("Sample data")
	mimeType := "text/plain"
	username := "user1@email.com"
	password := "password1"

	res := PostFragment(r, fileData, mimeType, username, password)

	w := httptest.NewRecorder()
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

func TestGetFragmentInfo(t *testing.T) {
	setup := PreTestSetup("debug")
	defer setup()

	r := getRouter()
	fileData := []byte("Sample data")
	mimeType := "text/plain"
	username := "user1@email.com"
	password := "password1"

	postFragmentResponse := PostFragment(r, fileData, mimeType, username, password)

	w := httptest.NewRecorder()
	fmt.Println("Location: ", postFragmentResponse.Location)
	getReq, _ := http.NewRequest("GET", postFragmentResponse.Location+"/info", nil)
	getReq.SetBasicAuth("user1@email.com", "password1")
	r.ServeHTTP(w, getReq)

	var getFragmentInfoResponse GetFragmentInfoResponse
	json.Unmarshal(w.Body.Bytes(), &getFragmentInfoResponse)

	assert.Equal(t, postFragmentResponse.Metadata, getFragmentInfoResponse.Fragment)
}

func TestGetConvertedFragment(t *testing.T) {
	setup := PreTestSetup()
	defer setup()

	r := getRouter()
	fileData := []byte("### Hello!\n")
	mimeType := "text/markdown"
	username := "user1@email.com"
	password := "password1"

	res := PostFragment(r, fileData, mimeType, username, password)

	w := httptest.NewRecorder()
	fmt.Println("Location: ", res.Location)
	getReq, _ := http.NewRequest("GET", res.Location+".html", nil)
	getReq.SetBasicAuth("user1@email.com", "password1")
	r.ServeHTTP(w, getReq)

	size, _ := strconv.Atoi(w.Result().Header.Get("Content-Length"))
	fmt.Println(size)
	retrievedFileData := w.Body.Bytes()

	assert.Equal(t, []byte("<h3>Hello!</h3>\n"), retrievedFileData)
}
func TestAwsAuthenticationEmptyAuthorizationHeader(t *testing.T) {
	setup := PreTestSetup("prod")
	defer setup()

	r := getRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/fragments", nil)
	req.Header.Add("Authorization", "")

	r.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Result().StatusCode)
}
