package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

	tempFile, _ := os.CreateTemp("", "*.txt")
	defer os.Remove(tempFile.Name())
	fileData := "Some test data in the file!"
	tempFile.Write([]byte(fileData))
	tempFile.Seek(0, 0)

	// Learned from https://andrew-mccall.com/blog/2024/06/golang-send-multipart-form-data-to-api-endpoint/
	buf := &bytes.Buffer{}
	mpw := multipart.NewWriter(buf)
	fwriter, _ := mpw.CreateFormFile("file", tempFile.Name())
	mpw.FormDataContentType()
	io.Copy(fwriter, tempFile)
	mpw.Close()

	req, _ := http.NewRequest("POST", "/v1/fragments", buf)
	req.Header.Add("Content-Type", mpw.FormDataContentType())
	req.SetBasicAuth("user1@email.com", "password1")

	_, header, _ := req.FormFile("file")
	header.Header.Set("Content-Type", "text/plain")

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
	tempFile, _ := os.Create("temp.txt")
	tempLocation, _ := filepath.Abs(tempFile.Name())
	fmt.Println("Current path: ", tempLocation)
	fileData := "Some test data in the file!"
	tempFile.Write([]byte(fileData))
	tempFile.Seek(0, 0)

	// Learned from https://andrew-mccall.com/blog/2024/06/golang-send-multipart-form-data-to-api-endpoint/
	buf := &bytes.Buffer{}
	mpw := multipart.NewWriter(buf)
	fwriter, _ := mpw.CreateFormFile("file", tempFile.Name())
	mpw.FormDataContentType()
	io.Copy(fwriter, tempFile)
	mpw.Close()

	req, _ := http.NewRequest("POST", "/v1/fragments", buf)
	req.Header.Add("Content-Type", mpw.FormDataContentType())
	req.SetBasicAuth("user1@email.com", "password1")

	_, header, _ := req.FormFile("file")
	header.Header.Set("Content-Type", "text/plain")

	r := getRouter()
	r.ServeHTTP(w, req)
	fmt.Println(GetBody(w.Body.Bytes()))
	location := GetBody(w.Body.Bytes())["Location"]

	w = httptest.NewRecorder()
	fmt.Println("Location: ", location)
	getReq, _ := http.NewRequest("GET", location, nil)
	getReq.SetBasicAuth("user1@email.com", "password1")
	r.ServeHTTP(w, getReq)

	size, _ := strconv.Atoi(w.Result().Header.Get("Content-Length"))
	fmt.Println(size)
	retrievedFileBuffer := make([]byte, size)
	w.Body.Read(retrievedFileBuffer)
	fmt.Println(string(retrievedFileBuffer))

	assert.Equal(t, fileData, string(retrievedFileBuffer))

	tempFile.Close()
	defer (func() {

		err := os.Remove(tempLocation)
		fmt.Println(err)
	})()
}
