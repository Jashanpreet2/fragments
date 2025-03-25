package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Jashanpreet2/fragments/localauthentication"
	"github.com/gin-gonic/gin"
	"github.com/gohugoio/hugo/common/hashing"
	cognitoJwtVerify "github.com/jhosan7/cognito-jwt-verify"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

var logger *zap.Logger
var sugar *zap.SugaredLogger
var localCsvAuthentication bool
var mode string

func Initialize() {
	var unspecifiedEnvironmentMessage string = "Server environment unspecified. Please specify debug or prod as the command line argument."
	if len(os.Args) < 2 {
		log.Fatal(unspecifiedEnvironmentMessage)
	}
	mode = strings.ToLower(os.Args[len(os.Args)-1])
	if mode != "debug" && mode != "prod" {
		log.Fatal(unspecifiedEnvironmentMessage)
	}

	// Logger setup
	// logger = getLogger(os.Getenv("LOG_LEVEL"))
	logger, _ = zap.NewDevelopment()

	sugar = logger.Sugar()

	// Load environment variables
	var err error
	if os.Getenv("TEST_PROFILE_PATH") == "" && mode == "debug" {
		err = godotenv.Load(".env.debug")
	} else if os.Getenv("AWS_COGNITO_POOL_ID") == "" && os.Getenv("AWS_COGNITO_CLIENT_ID") == "" && mode == "prod" {
		err = godotenv.Load(".env.prod")
	} else if os.Getenv("TEST_PROFILE_PATH") == "" && os.Getenv("AWS_COGNITO_POOL_ID") == "" && os.Getenv("AWS_COGNITO_CLIENT_ID") == "" {
		sugar.Fatal("Mode is neither debug nor prod. Ensure that the correct mode was passed when starting the application.")
	}

	if err != nil {
		log.Fatal("Failed to load environment variables")
	}

	if mode == "debug" {
		localCsvAuthentication = true
	} else {
		localCsvAuthentication = false
	}

	// Check that the necessary environment variables are present
	if os.Getenv("AWS_COGNITO_POOL_ID") == "" && os.Getenv("AWS_COGNITO_CLIENT_ID") == "" && !localCsvAuthentication {
		sugar.Fatal("Unable to find AWS_COGNITO_POOL_ID and AWS_COGNITO_CLIENT_ID")
	}
}

// Returns the specific logger based on the log level passed.
// func getLogger(logLevel string) *zap.Logger {
// 	if logLevel == "Debug" {
// 		return prettyconsole.NewLogger(zap.DebugLevel)
// 	} else if logLevel == "Info" {
// 		return prettyconsole.NewLogger(zap.InfoLevel)
// 	} else if logLevel == "Warn" {
// 		return prettyconsole.NewLogger(zap.WarnLevel)
// 	} else if logLevel == "Error" {
// 		return prettyconsole.NewLogger(zap.ErrorLevel)
// 	} else if logLevel == "Fatal" {
// 		return prettyconsole.NewLogger(zap.FatalLevel)
// 	} else {
// 		return prettyconsole.NewLogger(zap.InfoLevel)
// 	}
// }

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
	return func(c *gin.Context) {
		if localCsvAuthentication {
			sugar.Info("Local authentication using information from CSV files.")
			req := c.Request
			if username, password, ok := req.BasicAuth(); ok {
				sugar.Info("Ok so far")
				if localauthentication.AuthenticateTestProfile(os.Getenv("TEST_PROFILE_PATH"), username, password) {
					c.Set("username", username)
					c.Next()
				} else {
					c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid details"})
					c.Abort()
				}
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid details"})
				c.Abort()
			}
		} else {
			{
				cognitoConfig := cognitoJwtVerify.Config{
					UserPoolId: os.Getenv("AWS_COGNITO_POOL_ID"),
					ClientId:   os.Getenv("AWS_COGNITO_CLIENT_ID"),
					TokenUse:   "id",
				}

				// sugar.Info("Authentication in process")
				token := c.GetHeader("Authorization")

				if !strings.HasPrefix(token, "Bearer ") {
					c.JSON(http.StatusUnauthorized, gin.H{"message": "Unable to login"})
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
					sugar.Error("Failed to initialize cognito jwt verify")
					c.JSON(http.StatusInternalServerError, gin.H{"message": "System failed to start verification"})
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

				var v map[string]any
				if err := json.Unmarshal(jsonData, &v); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"message": "Unable to process request"})
					sugar.Error(err)
					sugar.Warn("Failed to retrieve user data from request body")
					c.Abort()
				}
				sugar.Info(v["cognito:username"])
				c.Set("username", v["cognito:username"])
				c.Next()
			}
		}

	}
}

func getRouter() *gin.Engine {
	Initialize()

	// Create router
	r := gin.New()
	r.MaxMultipartMemory = 8 << 23
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
		username := c.GetString("username")
		sugar.Info("Content type: ", c.GetHeader("Content-Type"))
		sugar.Info(username)
		if username == "" {
			sugar.Info("Request passed through authentication but still failed to retrieve the username")
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to parse username"})
		}
		fragmentIds := GetUserFragmentIds(hashing.HashString(username))

		if c.Query("expand") == "1" {
			fragments := []Fragment{}
			for _, fragmentId := range fragmentIds {
				fragment, found := GetFragment(hashing.HashString(username), fragmentId)
				if !found {
					sugar.Info("Failed to find fragment with fragment id " + fragmentId + " for user " + username)
				} else {
					fragments = append(fragments, fragment)
				}
				fragmentsjson, _ := json.Marshal(fragments)
				sugar.Info(string(fragmentsjson))
			}
			c.JSON(http.StatusOK, gin.H{"status": "ok", "fragments": fragments})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok", "fragment_ids": fragmentIds})
	})
	v1.POST("/fragments", func(c *gin.Context) {
		fileData, err := c.GetRawData()
		if err != nil {
			sugar.Info(err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Unable to retrieve file data from the request body!"})
		}
		username := hashing.HashString(c.GetString("username"))
		fragment_id := GenerateID(username)
		sugar.Info("DATAA:", string(fileData))
		fragmentType := c.GetHeader("Content-Type")
		if !IsSupportedType(fragmentType) {
			c.JSON(http.StatusUnsupportedMediaType, gin.H{"message": "The specified file format is currently not supported!"})
			sugar.Infof("User tried to store fragment of type %s", fragmentType)
			return
		}
		fragment := Fragment{strconv.Itoa(fragment_id), username, time.Now(), time.Now(), fragmentType, len(fileData)}
		fragment.SetData(fileData)
		sugar.Infof("File data being saved: %s", fileData)
		fragment.Save()

		scheme := "http://"
		if c.Request.TLS != nil {
			scheme = "https://"
		}

		c.Header("Location", scheme+c.Request.Host+fmt.Sprintf("/v1/fragment/%s", fragment.Id))
		c.JSON(http.StatusCreated, gin.H{"status": "ok", "message": "Fragment has successfully been saved",
			"fragment": fragment})
		// c.JSON(http.StatusOK, gin.H{"abc": "asja"})
		c.Abort()
	})
	v1.GET("/fragment/:id", func(c *gin.Context) {
		fragment_id := c.Param("id")
		var ext string
		for i := len(fragment_id) - 1; i > 0; i-- {
			if fragment_id[i] == '.' {
				ext = fragment_id[i:]
				fragment_id = fragment_id[0:i]
				break
			}
		}
		username := c.GetString("username")
		sugar.Infof("Request to fetch fragments. User ID: %s. Fragment_id: %s", username, fragment_id)
		fragment, ok := GetFragment(hashing.HashString(username), fragment_id)
		sugar.Info("ok?", ok)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"message": "Failed to find fragments with " +
				"the specified user id and fragment_id"})
			sugar.Error("Failed to find user's fragments. Check if the username was hashed successfully")
			return
		}
		sugar.Info(fragment.MimeType())
		var err error
		var fileData []byte
		var mimeType string
		if ext == "" {
			fileData, ok = fragment.GetData()
			mimeType = fragment.MimeType()
			if !ok {
				sugar.Info("Failed to find the fragment")
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to find the fragment data"})
				return
			}
		} else {
			fileData, mimeType, err = fragment.ConvertMimetype(ext)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
				return
			}
		}
		sugar.Info("Data in file: ", string(fileData))
		c.Header("Content-Length", strconv.Itoa(len(fileData)))
		c.Data(200, mimeType, fileData)
	})

	v1.GET("/fragment/:id/info", func(c *gin.Context) {
		id := c.Param("id")

		fragment, ok := GetFragment(hashing.HashString(c.GetString("username")), id)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"message": "Unable to find the specified fragment!"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok", "fragment": fragment})
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
		sugar.Fatal("Failed to start the server")
	}
}
