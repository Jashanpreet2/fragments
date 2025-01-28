package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	cognitoJwtVerify "github.com/jhosan7/cognito-jwt-verify"
	"github.com/joho/godotenv"
	prettyconsole "github.com/thessem/zap-prettyconsole"
	"go.uber.org/zap"
)

var logger *zap.Logger
var sugar *zap.SugaredLogger
var localCsvAuthentication bool

func Initialize() {
	var err error
	// Load environment variables
	err = godotenv.Load("env.test")
	if err != nil {
		log.Fatal("Failed to load environment variables")
	}

	// Logger setup
	logger = getLogger(os.Getenv("LOG_LEVEL"))
	defer logger.Sync()
	sugar = logger.Sugar()

	if os.Getenv("TEST_PROFILE_PATH") == "" {
		localCsvAuthentication = false
	} else {
		localCsvAuthentication = true
	}

	// Check that the necessary environment variables are present
	if os.Getenv("AWS_COGNITO_POOL_ID") == "" && os.Getenv("AWS_COGNITO_CLIENT_ID") == "" && !localCsvAuthentication {
		sugar.Fatal("Unable to find AWS_COGNITO_POOL_ID and AWS_COGNITO_CLIENT_ID")
	}

	if os.Getenv("AWS_COGNITO_POOL_ID") != "" && os.Getenv("AWS_COGNITO_CLIENT_ID") != "" && localCsvAuthentication {
		sugar.Fatal("Found both development and production environment variables")
	}
}

// Returns the specific logger based on the log level passed.
func getLogger(logLevel string) *zap.Logger {
	if logLevel == "Debug" {
		return prettyconsole.NewLogger(zap.DebugLevel)
	} else if logLevel == "Info" {
		return prettyconsole.NewLogger(zap.InfoLevel)
	} else if logLevel == "Warn" {
		return prettyconsole.NewLogger(zap.WarnLevel)
	} else if logLevel == "Error" {
		return prettyconsole.NewLogger(zap.ErrorLevel)
	} else if logLevel == "Fatal" {
		return prettyconsole.NewLogger(zap.FatalLevel)
	} else {
		return prettyconsole.NewLogger(zap.InfoLevel)
	}
}

// Enables Cors and adds any other relevant headers including Cache-Control
// https://stackoverflow.com/a/29439630
func SetHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")
		c.Writer.Header().Set("Cache-Control", "no-cache")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func authenticate() gin.HandlerFunc {
	if localCsvAuthentication {
		return func(c *gin.Context) {
			fmt.Print("LOCAL AUTHANETICA")
			req := c.Request
			if username, password, ok := req.BasicAuth(); ok {
				if AuthenticateTestProfile(os.Getenv("TEST_PROFILE_PATH"), username, password) {
					c.Next()
				} else {
					c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid details"})
					c.Abort()
				}
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid details"})
				c.Abort()
			}
		}
	} else {
		return func(c *gin.Context) {
			cognitoConfig := cognitoJwtVerify.Config{
				UserPoolId: os.Getenv("AWS_COGNITO_POOL_ID"),
				ClientId:   os.Getenv("AWS_COGNITO_CLIENT_ID"),
				TokenUse:   "id",
			}

			// sugar.Info("Authentication in process")
			token := c.GetHeader("Authorization")

			if !strings.HasPrefix(token, "Bearer ") {
				c.JSON(http.StatusBadRequest, gin.H{"message": "Unable to login"})
				c.Abort()
				return
			}

			token = token[len("Bearer "):]

			if token == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "Unable to login"})
				c.Abort()
				return
			}

			verify, err := cognitoJwtVerify.Create(cognitoConfig)
			if err != nil {
				sugar.Fatal("Failed to initialize cognito jwt verify")
				c.Abort()
				return
			}

			payload, err := verify.Verify(token)
			if err != nil {
				sugar.Info("Failed to verify token", zap.Error(err))
				c.JSON(http.StatusUnauthorized, gin.H{"message": "Unable to login"})
				c.Abort()
				return
			}

			jsonData, err := json.Marshal(payload)
			if err != nil {
				sugar.Info("Failed to parse token", zap.Error(err))
				c.JSON(http.StatusUnauthorized, gin.H{"message": "Unable to login"})
				c.Abort()
				return
			}

			sugar.Info(string(jsonData))
			c.Next()
		}
	}

}

func getRouter() *gin.Engine {
	Initialize()

	// Create router
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(SetHeaders())
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok",
			"author":    "Jashanpreet Singh",
			"githuburl": "https://github.com/Jashanpreet2/fragments",
			"version":   "1"})
	})

	v1 := r.Group("v1")
	v1.Use(authenticate())
	v1.GET("/fragments", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "You have been verified!"})
	})

	return r
}

func main() {
	// Create and assign logger instance to the global variable
	var err error

	Initialize()

	// Start server
	port := os.Getenv("PORT")
	r := getRouter()
	err = r.Run(port)
	if err != nil {
		// sugar.Fatal("Failed to start the server")
	}
}
