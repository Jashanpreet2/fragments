package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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
	if mode == "debug" {
		err = godotenv.Load(".env.debug")
	} else if mode == "prod" {
		err = godotenv.Load(".env.prod")
	} else {
		sugar.Fatal("Mode is neither debug nor prod. Ensure that the correct mode was passed when starting the application.")
	}

	if err != nil {
		log.Fatal("Failed to load environment variables")
	}

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

				sugar.Info(jsonData)
				var v map[string]string
				if err := json.Unmarshal(jsonData, &v); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"message": "Unable to process request"})
					sugar.Warn("Failed to retrieve user data from request body")
					c.Abort()
				}
				sugar.Info(v["cognito:username"])
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
		username, ok := GetUsername(c, localCsvAuthentication)
		if !ok {
			sugar.Info("Request passed through authentication but still failed to retrieve the username")
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to parse username"})
		}
		fragment_ids := GetUserFragmentIds(hashing.HashString(username))
		c.JSON(http.StatusOK, gin.H{"status": "ok", "fragment_ids": fragment_ids})
	})
	v1.POST("/fragments", func(c *gin.Context) {
		fileHeader, _ := c.FormFile("file")
		username, ok := GetUsername(c, localCsvAuthentication)
		if !ok {
			sugar.Error("Failed to parse username!")
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to parse username"})
			return
		}
		file, err := fileHeader.Open()
		if err != nil {
			sugar.Error("Failed to read file!")
		}
		fragment_id := GenerateID(username)
		fileData := make([]byte, 512)
		if _, err := file.Read(fileData); err != nil {
			sugar.Error("Failed to read user file")
			c.JSON(http.StatusBadRequest, gin.H{"message": "Unable to process the uploaded file"})
		}
		if _, err := file.Seek(0, 0); err != nil {
			sugar.Error("Failed to seek file back to the start")
			c.JSON(http.StatusBadRequest, gin.H{"message": "Server failed in processing the file"})
		}
		sugar.Info("DATAA:", string(fileData))
		fragmentType := fileHeader.Header.Get("Content-Type")
		if !IsSupportedType(fragmentType) {
			c.JSON(http.StatusBadRequest, gin.H{"message": "The specified file format is currently not supported!"})
			sugar.Infof("User tried to store fragment of type %s", fragmentType)
			return
		}
		fragment := Fragment{strconv.Itoa(fragment_id), hashing.HashString(username), time.Now(), time.Now(), fragmentType, fileHeader.Size, fileHeader.Filename}
		fragment.SetData(file)
		fragment.Save()

		metadata, _ := fragment.GetJson()

		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Fragment has successfully been saved",
			"Location": c.Request.Host + fmt.Sprintf("/v1/fragment/%s", fragment.Id),
			"metadata": metadata})
	})
	v1.GET("/fragment/:id", func(c *gin.Context) {
		sugar.Infof("Request to fetch fragments. User ID: %s. Fragment_id: %s", c.Query("userid"), c.Query("fragment_id"))
		fragment_id := c.Param("id")
		userid, _ := GetUsername(c, localCsvAuthentication)
		fragment, ok := GetFragment(hashing.HashString(userid), fragment_id)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to find fragments with " +
				"the specified user id and fragment_id"})
			sugar.Error("Failed to find user's fragments. Check if the username was hashed successfully")
			return
		}
		file, ok := fragment.GetData()
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to find the fragment data"})
			return
		}
		pwd, err := os.Getwd()
		if err != nil {
			sugar.Info(err)
		}
		new_file, err := os.Create(filepath.Join(pwd, "tmp"+filepath.Base(fragment.FragmentName)))
		if err != nil {
			sugar.Info(err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Server failed to send file"})
			return
		}
		fmt.Println("filepath", new_file.Name())
		io.Copy(new_file, file)
		c.FileAttachment(new_file.Name(), fragment.FragmentName)

		new_file.Close()
		os.Remove(new_file.Name())
	})

	return r
}

func main() {
	// Create and assign logger instance to the global variable
	var err error

	Initialize()

	fmt.Println(filepath.Join("files", "aja"))

	// Start server
	port := os.Getenv("PORT")
	r := getRouter()
	err = r.Run(port)
	if err != nil {
		sugar.Fatal("Failed to start the server")
	}
}
