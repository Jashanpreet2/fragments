package config

import (
	"os"
	"strings"

	"github.com/Jashanpreet2/fragments/internal/logger"
	"github.com/joho/godotenv"
)

var LocalCsvAuthentication bool
var loaded bool

func Config() {
	if loaded {
		return
	}

	var unspecifiedEnvironmentMessage string = "Server environment unspecified. Please specify debug or prod as the command line argument."
	if len(os.Args) < 2 {
		logger.Sugar.Fatal(unspecifiedEnvironmentMessage)
	}
	mode := strings.ToLower(os.Args[len(os.Args)-1])
	if mode != "debug" && mode != "prod" {
		logger.Sugar.Fatal(unspecifiedEnvironmentMessage)
	}

	// Load environment variables
	var err error
	if os.Getenv("TEST_PROFILE_PATH") == "" && mode == "debug" {
		err = godotenv.Load(".env.debug")
	} else if os.Getenv("AWS_COGNITO_POOL_ID") == "" && os.Getenv("AWS_COGNITO_CLIENT_ID") == "" && mode == "prod" {
		err = godotenv.Load(".env.prod")
	} else if os.Getenv("TEST_PROFILE_PATH") == "" && os.Getenv("AWS_COGNITO_POOL_ID") == "" && os.Getenv("AWS_COGNITO_CLIENT_ID") == "" {
		logger.Sugar.Fatal("Mode is neither debug nor prod. Ensure that the correct mode was passed when starting the application.")
	}

	if err != nil {
		logger.Sugar.Fatal(err)
	}

	if os.Getenv("TEST_PROFILE_PATH") == "" {
		LocalCsvAuthentication = false
	} else {
		LocalCsvAuthentication = true
	}

	// Check that the necessary environment variables are present
	if os.Getenv("AWS_COGNITO_POOL_ID") == "" && os.Getenv("AWS_COGNITO_CLIENT_ID") == "" && !LocalCsvAuthentication {
		logger.Sugar.Fatal("Unable to find AWS_COGNITO_POOL_ID and AWS_COGNITO_CLIENT_ID")
	}

	loaded = true
}
