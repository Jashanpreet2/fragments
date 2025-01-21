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

// Cors middleware enabling code from
// https://stackoverflow.com/a/29439630
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func authenticate() gin.HandlerFunc {
	fmt.Println("In authentication")
	return func(c *gin.Context) {
		cognitoConfig := cognitoJwtVerify.Config{
			UserPoolId: os.Getenv("AWS_COGNITO_POOL_ID"),
			ClientId:   os.Getenv("AWS_COGNITO_CLIENT_ID"),
			TokenUse:   "id",
		}

		sugar.Info("Authentication in process")
		token := c.GetHeader("Authorization")

		if !strings.HasPrefix(token, "Bearer ") {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Unable to login"})
			c.Abort()
			return
		}

		token = token[len("Bearer "):]

		if token == "" {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Unable to login"})
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
			c.JSON(http.StatusBadRequest, gin.H{"message": "Unable to login"})
			c.Abort()
			return
		}

		jsonData, err := json.Marshal(payload)
		if err != nil {
			sugar.Info("Failed to parse token", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"message": "Unable to login"})
			c.Abort()
			return
		}

		sugar.Info(string(jsonData))
		c.Next()
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	logger.Info("Received a request!")
	fmt.Fprintf(w, "Sample response")
}

func main() {
	// Create and assign logger instance to the global variable
	var err error
	logger = prettyconsole.NewLogger(zap.DebugLevel)
	sugar = logger.Sugar()

	// Load environment variables
	err = godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load environment variables")
	}

	// Create router
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(CORSMiddleware())
	port := os.Getenv("PORT")

	v1 := r.Group("v1")
	v1.Use(authenticate())
	v1.GET("/fragments", func(c *gin.Context) {
		c.JSON(http.StatusAccepted, gin.H{"message": "You have been verified!"})
	})

	r.Run(port)
}
