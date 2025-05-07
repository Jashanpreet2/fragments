package testutils

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/Jashanpreet2/fragments/internal/config"
	"github.com/Jashanpreet2/fragments/internal/fragment"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var modeAdded = false

func PreTestSetup(mode string) func() {
	if !modeAdded {
		os.Args = append(os.Args, mode)
		modeAdded = true
	} else {
		os.Args[len(os.Args)-1] = mode
	}
	if mode == "debug" {
		godotenv.Load("../../")
	}
	config.Config()
	return func() {
		fragment.ResetDB()
	}
}

func CreateTestFragment() fragment.Fragment {
	return fragment.Fragment{Id: "1", OwnerId: "user", Created: time.Now(), Updated: time.Now(), FragmentType: "text", Size: 5}
}

func PostFragment(r *gin.Engine, data []byte, mimeType string, username string, password string) *httptest.ResponseRecorder {
	// Set up and make request
	w := httptest.NewRecorder()
	fileData := []byte(data)

	req, _ := http.NewRequest("POST", "/v1/fragments", bytes.NewReader(fileData))
	req.Header.Add("Content-Type", mimeType)
	req.SetBasicAuth(username, password)

	r.ServeHTTP(w, req)

	return w
}
