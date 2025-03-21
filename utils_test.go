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

func PreTestSetup() func() {
	os.Args = append(os.Args, "debug")
	Initialize()
	return func() {
		// Might have some test teardown logic later on
	}
}

func CreateTestFragment() Fragment {
	return Fragment{"1", "user", time.Now(), time.Now(), "text", 5}
}

