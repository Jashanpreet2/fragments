package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

var modeAdded = false

func PreTestSetup(mode string) func() {
	os.Unsetenv("TEST_PROFILE_PATH")
	os.Unsetenv("AWS_COGNITO_POOL_ID")
	os.Unsetenv("AWS_COGNITO_CLIENT_ID")
	if !modeAdded {
		os.Args = append(os.Args, mode)
		modeAdded = true
	} else {
		os.Args[len(os.Args)-1] = mode
	}
	Initialize()
	return func() {
		// Might have some test teardown logic later on
	}
}

func CreateTestFragment() Fragment {
	return Fragment{"1", "user", time.Now(), time.Now(), "text", 5}
}

func PostFragment(r *gin.Engine, data []byte, mimeType string, username string, password string) PostFragmentResponse {
	// Set up and make request
	w := httptest.NewRecorder()
	fileData := []byte(data)

	req, _ := http.NewRequest("POST", "/v1/fragments", bytes.NewReader(fileData))
	req.Header.Add("Content-Type", mimeType)
	req.SetBasicAuth(username, password)

	r.ServeHTTP(w, req)

	var res PostFragmentResponse
	json.Unmarshal(w.Body.Bytes(), &res)

	return res
}
