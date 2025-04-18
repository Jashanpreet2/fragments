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

	"github.com/Jashanpreet2/fragments/internal/testutils"
	"github.com/Jashanpreet2/fragments/internal/utils"
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
	Message  string
	Fragment FragmentMetadata
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
	setup := testutils.PreTestSetup("debug")
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
	setup := testutils.PreTestSetup("debug")
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
	setup := testutils.PreTestSetup("debug")
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
	jsonMap := utils.GetBody(bodyBytes)

	// Assert correct response
	assert.Equal(t, "ok", jsonMap["status"])
}

func TestBodyInformation(t *testing.T) {
	setup := testutils.PreTestSetup("debug")
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
	jsonMap := utils.GetBody(bodyBytes)

	// Assert correct response
	assert.Equal(t, "https://github.com/Jashanpreet2/fragments", jsonMap["githuburl"])
	assert.Equal(t, "Jashanpreet Singh", jsonMap["author"])
	assert.Equal(t, "1", jsonMap["version"])
}

func TestUnauthenticatedRequest(t *testing.T) {
	setup := testutils.PreTestSetup("debug")
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
	setup := testutils.PreTestSetup("debug")
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
	setup := testutils.PreTestSetup("debug")
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
	setup := testutils.PreTestSetup("debug")
	defer setup()

	// Set up and make request
	w := httptest.NewRecorder()

	fileData := []byte("Some test data in the file!")

	req, _ := http.NewRequest("POST", "/v1/fragments", bytes.NewReader(fileData))
	req.Header.Add("Content-Type", "text/plain")
	req.SetBasicAuth("user1@email.com", "password1")

	r := getRouter()
	r.ServeHTTP(w, req)

	fmt.Println(utils.GetBody(w.Body.Bytes()))

	// Assert
	assert.Equal(t, 201, w.Result().StatusCode)
}

func TestGetFragments(t *testing.T) {
	setup := testutils.PreTestSetup("debug")
	defer setup()

	r := getRouter()
	fileData := []byte("Sample data")
	mimeType := "text/plain"
	username := "user1@email.com"
	password := "password1"

	w := testutils.PostFragment(r, fileData, mimeType, username, password)
	var postFragmentResponse PostFragmentResponse
	json.Unmarshal(w.Body.Bytes(), &postFragmentResponse)

	req, _ := http.NewRequest("GET", "/v1/fragments", nil)
	req.SetBasicAuth(username, password)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var response GetFragmentsResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
}

func TestGetFragmentsExpanded(t *testing.T) {
	setup := testutils.PreTestSetup("debug")
	defer setup()

	r := getRouter()
	fileData := []byte("Sample data")
	mimeType := "text/plain"
	username := "user1@email.com"
	password := "password1"

	w := testutils.PostFragment(r, fileData, mimeType, username, password)
	var postFragmentResponse PostFragmentResponse
	json.Unmarshal(w.Body.Bytes(), &postFragmentResponse)

	req, _ := http.NewRequest("GET", "/v1/fragments?expand=1", nil)
	req.SetBasicAuth(username, password)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	// var response GetFragmentsExpandedResponse
	// json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
}

func TestGetFragment(t *testing.T) {
	setup := testutils.PreTestSetup("debug")
	defer setup()

	r := getRouter()
	fileData := []byte("Sample data")
	mimeType := "text/plain"
	username := "user1@email.com"
	password := "password1"

	w := testutils.PostFragment(r, fileData, mimeType, username, password)
	location := w.Header().Get("Location")

	fmt.Println("Location: ", location)
	getReq, _ := http.NewRequest("GET", location, nil)
	getReq.SetBasicAuth("user1@email.com", "password1")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, getReq)

	size, _ := strconv.Atoi(w.Result().Header.Get("Content-Length"))
	fmt.Println(size)
	retrievedFileBuffer := make([]byte, size)
	w.Body.Read(retrievedFileBuffer)
	fmt.Println(string(retrievedFileBuffer))

	assert.Equal(t, fileData, retrievedFileBuffer)
}

func TestGetNonExistentFragment(t *testing.T) {
	setup := testutils.PreTestSetup("debug")
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
	setup := testutils.PreTestSetup("debug")
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
	setup := testutils.PreTestSetup("debug")
	defer setup()

	r := getRouter()
	fileData := []byte("Sample data")
	mimeType := "text/plain"
	username := "user1@email.com"
	password := "password1"

	w := testutils.PostFragment(r, fileData, mimeType, username, password)
	var postFragmentResponse PostFragmentResponse
	json.Unmarshal(w.Body.Bytes(), &postFragmentResponse)
	location := w.Header().Get("Location")

	fmt.Println("Location: ", location)
	getReq, _ := http.NewRequest("GET", location+"/info", nil)
	getReq.SetBasicAuth("user1@email.com", "password1")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, getReq)

	var getFragmentInfoResponse GetFragmentInfoResponse
	json.Unmarshal(w.Body.Bytes(), &getFragmentInfoResponse)

	assert.Equal(t, postFragmentResponse.Fragment, getFragmentInfoResponse.Fragment)
}

func TestGetConvertedFragment(t *testing.T) {
	setup := testutils.PreTestSetup("debug")
	defer setup()

	r := getRouter()
	fileData := []byte("### Hello!\n")
	mimeType := "text/markdown"
	username := "user1@email.com"
	password := "password1"

	w := testutils.PostFragment(r, fileData, mimeType, username, password)
	location := w.Header().Get("Location")
	fmt.Println("Location: ", location)
	getReq, _ := http.NewRequest("GET", location+".html", nil)
	getReq.SetBasicAuth("user1@email.com", "password1")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, getReq)

	size, _ := strconv.Atoi(w.Result().Header.Get("Content-Length"))
	fmt.Println(size)
	retrievedFileData := w.Body.Bytes()

	assert.Equal(t, []byte("<h3>Hello!</h3>\n"), retrievedFileData)
}

func TestGetConvertedFragmentInvalidExtension(t *testing.T) {
	setup := testutils.PreTestSetup("debug")
	defer setup()

	r := getRouter()
	fileData := []byte("### Hello!\n")
	mimeType := "text/markdown"
	username := "user1@email.com"
	password := "password1"

	w := testutils.PostFragment(r, fileData, mimeType, username, password)
	location := w.Header().Get("Location")

	fmt.Println("Location: ", w.Header().Get("Location"))
	getReq, _ := http.NewRequest("GET", location+".invalidextension", nil)
	getReq.SetBasicAuth("user1@email.com", "password1")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, getReq)

	assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
}

// func TestAwsAuthenticationValidAuthorization(t *testing.T) {
// 	setup :=testutils.PreTestSetup("prod")
// 	defer setup()

// 	r := getRouter()

// 	w := httptest.NewRecorder()
// 	req, _ := http.NewRequest("GET", "/v1/fragments", nil)
// 	req.Header.Add("Authorization", "Bearer eyJraWQiOiJIYUFZbXFiUUlJXC8xanpRYUczcUswdjdlQnhiNU02SFwvSlZMRVJZY3I4Q2s9IiwiYWxnIjoiUlMyNTYifQ.eyJhdF9oYXNoIjoia3p3UmJOejA3N1RfaXRZeFRVQ0ozQSIsInN1YiI6IjQ0Nzg2NDU4LTIwNjEtNzA0OC05YWFlLWE5Mzg3Njk1NmQ2NSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJpc3MiOiJodHRwczpcL1wvY29nbml0by1pZHAudXMtZWFzdC0xLmFtYXpvbmF3cy5jb21cL3VzLWVhc3QtMV9Fck9GZlhuWDQiLCJjb2duaXRvOnVzZXJuYW1lIjoiamFzaGFuMSIsIm9yaWdpbl9qdGkiOiJmM2RiNzQ4NS1lNDU1LTQ3NjgtYjVmOS0zMGY2Yzg2YTI2ODMiLCJhdWQiOiI1cWlndHU3NmF1MnM1cjM5bjI0amh0c3Y2IiwiZXZlbnRfaWQiOiJjNDQ0MTMxOS00NWU1LTQxMWQtYTVjYy1lNjg2NmM5NmNjZGUiLCJ0b2tlbl91c2UiOiJpZCIsImF1dGhfdGltZSI6MTc0MzQ3MTcwNSwiZXhwIjoxNzQzNDc1MzA1LCJpYXQiOjE3NDM0NzE3MDYsImp0aSI6IjNlNWRmYjhmLTg4MzMtNDljYS1iNWQxLTQyMmE3YWEyMzdhMyIsImVtYWlsIjoianNpbmdoMTAwOUBteXNlbmVjYS5jYSJ9.LI-yCVroCzEyMaSbyEV_wttqtaWzyQ67IPeeelsck3V-MpVjkLzmgmArw0SvyX_zPKgvlXbDZYlUHwYTFfo65kqyC8rFWbe_kwEu4IArOUYX-bfaZxgDFBR0s2KGgZ7zuAk72E4IypwPtyyjPDC7JinqypYEETgOv-1ohEByTMDY0FWj3434TMYSIYKo4xIhZKrfsTBy_q0Ai7UnMuFQlSNDn7ROBYqOgJBRjpaGiPt9A8_B_XlVlS9uhj5R2M8HSTIEI_B4pq6u6dOh_cjENlSPT8OK8P2pxGON_6Ni6rHU4b_sh0UgLDgvQWVwwLqsHTmZ6-yZQ0m07ybmbvV3Jw")

// 	r.ServeHTTP(w, req)
// 	var response GetFragmentsResponse
// 	json.Unmarshal(w.Body.Bytes(), &response)

// 	assert.Equal(t, "ok", response.Status)
// }

func TestAwsAuthenticationEmptyToken(t *testing.T) {
	setup := testutils.PreTestSetup("prod")
	defer setup()

	r := getRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/fragments", nil)
	req.Header.Add("Authorization", "Bearer ")

	r.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Result().StatusCode)
}

func TestAwsAuthenticationInvalidAuthorization(t *testing.T) {
	setup := testutils.PreTestSetup("prod")
	defer setup()

	r := getRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/fragments", nil)
	req.Header.Add("Authorization", "Bearer Invalid bearer token")
	r.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Result().StatusCode)

}

func TestAwsAuthenticationEmptyAuthorizationHeader(t *testing.T) {
	setup := testutils.PreTestSetup("prod")
	defer setup()

	r := getRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/fragments", nil)
	req.Header.Add("Authorization", "")

	r.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Result().StatusCode)
}

func TestPostInvalidMimetypeFragment(t *testing.T) {
	setup := testutils.PreTestSetup("debug")
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

	fmt.Println(utils.GetBody(w.Body.Bytes()))

	// Assert
	assert.Equal(t, http.StatusUnsupportedMediaType, w.Result().StatusCode)
}
