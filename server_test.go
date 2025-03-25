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
	setup := PreTestSetup("debug")
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
	setup := PreTestSetup("debug")
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
	setup := PreTestSetup("debug")
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
	setup := PreTestSetup("debug")
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
	setup := PreTestSetup("debug")
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
	setup := PreTestSetup("debug")
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
	setup := PreTestSetup("debug")
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
	setup := PreTestSetup("debug")
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
	assert.Equal(t, 201, w.Result().StatusCode)
}

func TestGetFragments(t *testing.T) {
	setup := PreTestSetup("debug")
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
	setup := PreTestSetup("debug")
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
	setup := PreTestSetup("debug")
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

func TestGetNonExistentFragment(t *testing.T) {
	setup := PreTestSetup("debug")
	defer setup()

	r := getRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/fragment/invalidid", nil)
	req.SetBasicAuth("user1@email.com", "password1")
	r.ServeHTTP(w, req)
	fmt.Print("Getting non fragment", w.Result().StatusCode)
	assert.Equal(t, 404, w.Result().StatusCode)
}

func TestGetNonExistentFragmentInfo(t *testing.T) {
	setup := PreTestSetup("debug")
	defer setup()

	r := getRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/fragment/invalidid/info", nil)
	req.SetBasicAuth("user1@email.com", "password1")
	r.ServeHTTP(w, req)
	fmt.Print("Getting non fragment", w.Result().StatusCode)
	assert.Equal(t, 404, w.Result().StatusCode)
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
	setup := PreTestSetup("debug")
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

func TestGetConvertedFragmentInvalidExtension(t *testing.T) {
	setup := PreTestSetup("debug")
	defer setup()

	r := getRouter()
	fileData := []byte("### Hello!\n")
	mimeType := "text/markdown"
	username := "user1@email.com"
	password := "password1"

	res := PostFragment(r, fileData, mimeType, username, password)

	w := httptest.NewRecorder()
	fmt.Println("Location: ", res.Location)
	getReq, _ := http.NewRequest("GET", res.Location+".invalidextension", nil)
	getReq.SetBasicAuth("user1@email.com", "password1")
	r.ServeHTTP(w, getReq)

	assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
}

func TestAwsAuthenticationValidAuthorization(t *testing.T) {
	setup := PreTestSetup("prod")
	defer setup()

	r := getRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/fragments", nil)
	req.Header.Add("Authorization", "Bearer eyJraWQiOiJIYUFZbXFiUUlJXC8xanpRYUczcUswdjdlQnhiNU02SFwvSlZMRVJZY3I4Q2s9IiwiYWxnIjoiUlMyNTYifQ.eyJhdF9oYXNoIjoiVlM1cllqSzcwZU1nQ1hBR0pKOC1qQSIsInN1YiI6IjQ0Nzg2NDU4LTIwNjEtNzA0OC05YWFlLWE5Mzg3Njk1NmQ2NSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJpc3MiOiJodHRwczpcL1wvY29nbml0by1pZHAudXMtZWFzdC0xLmFtYXpvbmF3cy5jb21cL3VzLWVhc3QtMV9Fck9GZlhuWDQiLCJjb2duaXRvOnVzZXJuYW1lIjoiamFzaGFuMSIsIm9yaWdpbl9qdGkiOiI1ZjNmZjY0MS0yZDJlLTRhMGMtOTZiYy04ZGU2NzcwNTg1YmUiLCJhdWQiOiI1cWlndHU3NmF1MnM1cjM5bjI0amh0c3Y2IiwiZXZlbnRfaWQiOiI4MGU3MzcxYy1hMmIyLTRmOTYtYTJiYi0wZmM1NTVhY2ExMWQiLCJ0b2tlbl91c2UiOiJpZCIsImF1dGhfdGltZSI6MTc0MjYyOTM5NCwiZXhwIjoxNzQyNjMyOTk0LCJpYXQiOjE3NDI2MjkzOTQsImp0aSI6ImYyNjBlZjU3LThmZWEtNGU0Yy1hZTUzLTc3YjJkYmEwMTBlMCIsImVtYWlsIjoianNpbmdoMTAwOUBteXNlbmVjYS5jYSJ9.O_I-jXY1iH1seuYXak7M0fbi8q_6zfJDyxRaSjsYt9S_5j6GiOlXjzgRbiMdrQa5wEk7uvzX0uvP1cAxI9MRKLzj21pZVDs5twpNjOyNLz_ZeCdLf1Xb8pXX_nAYYM9W3j3xoKWWpwg3xqyPqZy9hqxkkuCb0fP9JCYYuDUy3bbAM-IrF_5ySBSKbZLxl4jgUIt0-fvrNJpO6wF7Lmkk80XVoRdjwYahRorGVrydKJ5EsFq25jkRKILsJg_LxcKVg8BqwpopzGI79jZT20EPDOvTeNRX8xWqY9U-Dv80XzbWZMCxNuVxq9chqIHNtVzinYfTpN8k6CLCM4DxaK9mWA")

	r.ServeHTTP(w, req)
	var response GetFragmentsResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "ok", response.Status)
}

func TestAwsAuthenticationEmptyToken(t *testing.T) {
	setup := PreTestSetup("prod")
	defer setup()

	r := getRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/fragments", nil)
	req.Header.Add("Authorization", "Bearer ")

	r.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Result().StatusCode)
}

func TestAwsAuthenticationInvalidAuthorization(t *testing.T) {
	setup := PreTestSetup("prod")
	defer setup()

	r := getRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/fragments", nil)
	req.Header.Add("Authorization", "Bearer Invalid bearer token")
	r.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Result().StatusCode)

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

func TestPostInvalidMimetypeFragment(t *testing.T) {
	setup := PreTestSetup("debug")
	defer setup()

	// Set up and make request
	w := httptest.NewRecorder()

	fileData := []byte("Some test data in the file!")

	req, _ := http.NewRequest("POST", "/v1/fragments", bytes.NewReader(fileData))

	// Invalid mime-type
	req.Header.Add("Content-Type", "invalid/invalid")
	req.SetBasicAuth("user1@email.com", "password1")

	r := getRouter()
	r.ServeHTTP(w, req)

	fmt.Println(GetBody(w.Body.Bytes()))

	// Assert
	assert.Equal(t, 400, w.Result().StatusCode)
}
