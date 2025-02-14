package main

import (
	"os"
	"time"
)

func PreTestSetup() func() {
	os.Args = append(os.Args, "debug")
	Initialize()
	return func() {
		// Might have some test teardown logic later on
		return
	}
}

func CreateTestFragment() Fragment {
	return Fragment{"1", "user", time.Now(), time.Now(), "text", 5, "test fragment.type"}
}
