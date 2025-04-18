package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Jashanpreet2/fragments/internal/config"
	"github.com/Jashanpreet2/fragments/internal/fragment"
	"github.com/Jashanpreet2/fragments/internal/logger"
	"github.com/Jashanpreet2/fragments/localauthentication"
	"github.com/gin-gonic/gin"
	"github.com/gohugoio/hugo/common/hashing"
	cognitoJwtVerify "github.com/jhosan7/cognito-jwt-verify"
	"go.uber.org/zap"
)

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
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
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
		if config.LocalCsvAuthentication {
			logger.Sugar.Info("Local authentication using information from CSV files.")
			req := c.Request
			if username, password, ok := req.BasicAuth(); ok {
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

				// logger.Sugar.Info("Authentication in process")
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
					logger.Sugar.Error("Failed to initialize cognito jwt verify")
					c.JSON(http.StatusInternalServerError, gin.H{"message": "System failed to start verification"})
					c.Abort()
					return
				}

				payload, err := verify.Verify(token)
				if err != nil {
					logger.Sugar.Info("Failed to verify token", zap.Error(err))
					c.JSON(http.StatusUnauthorized, gin.H{"message": "Unable to login"})
					c.Abort()
					return
				}

				jsonData, err := json.Marshal(payload)
				if err != nil {
					logger.Sugar.Info("Failed to parse token", zap.Error(err))
					c.JSON(http.StatusUnauthorized, gin.H{"message": "Unable to login"})
					c.Abort()
					return
				}

				var v map[string]any
				if err := json.Unmarshal(jsonData, &v); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"message": "Unable to process request"})
					logger.Sugar.Error(err)
					logger.Sugar.Warn("Failed to retrieve user data from request body")
					c.Abort()
				}
				logger.Sugar.Info(v["cognito:username"])
				c.Set("username", v["cognito:username"])
				c.Next()
			}
		}
	}
}

func getRouter() *gin.Engine {
	// Create router
	r := gin.New()
	r.MaxMultipartMemory = 8 << 23
	r.Use(gin.Logger())
	r.Use(SetHeaders())
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok",
			"author":    "Jashanpreet Singh",
			"githuburl": "https://github.com/Jashanpreet2/fragments",
			"version":   "1",
			"hostname":  c.Request.Host})
	})

	v1 := r.Group("v1")
	v1.Use(authenticate())

	v1.GET("/fragments", func(c *gin.Context) {
		username := c.GetString("username")
		logger.Sugar.Info("Content type: ", c.GetHeader("Content-Type"))
		logger.Sugar.Info(username)
		if username == "" {
			logger.Sugar.Info("Request passed through authentication but still failed to retrieve the username")
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to parse username"})
		}
		fragmentIds, err := fragment.GetUserFragmentIds(hashing.HashString(username))
		if err != nil {
			logger.Sugar.Info(err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to retrieve fragment IDs"})
			return
		}
		logger.Sugar.Info(fragmentIds)
		if c.Query("expand") == "1" {
			fragments := []*fragment.Fragment{}
			for _, fragmentId := range fragmentIds {
				fragment, err := fragment.GetFragment(hashing.HashString(username), fragmentId)
				if err != nil {
					logger.Sugar.Info("Failed to find fragment with fragment id " + fragmentId + " for user " + username)
				} else {
					fragments = append(fragments, fragment)
				}
				fragmentsjson, _ := json.Marshal(fragments)
				logger.Sugar.Info(string(fragmentsjson))
			}
			c.JSON(http.StatusOK, gin.H{"status": "ok", "fragments": fragments})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok", "fragment_ids": fragmentIds})
	})
	v1.POST("/fragments", func(c *gin.Context) {
		fileData, err := c.GetRawData()
		if err != nil {
			logger.Sugar.Info(err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Unable to retrieve file data from the request body!"})
		}
		username := hashing.HashString(c.GetString("username"))
		fragment_id := fragment.GenerateID()
		logger.Sugar.Info("File data: ", string(fileData))
		fragmentType := c.GetHeader("Content-Type")
		if !fragment.IsSupportedType(fragmentType) {
			c.JSON(http.StatusUnsupportedMediaType, gin.H{"message": "The specified file format is currently not supported!"})
			logger.Sugar.Infof("User tried to store fragment of type %s", fragmentType)
			return
		}
		fragment := fragment.Fragment{
			Id:      fragment_id,
			OwnerId: username, Created: time.Now(),
			Updated:      time.Now(),
			FragmentType: fragmentType,
			Size:         len(fileData)}
		fragment.SetData(fileData)
		logger.Sugar.Infof("File data being saved: %s", fileData)
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
	v1.PUT("/fragments/:id", func(c *gin.Context) {
		fragment_id := c.Param("id")
		fileData, err := c.GetRawData()
		if err != nil {
			logger.Sugar.Info(err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Unable to retrieve file data from the request body!"})
		}
		username := hashing.HashString(c.GetString("username"))
		logger.Sugar.Info("File data: ", string(fileData))
		fragmentType := c.GetHeader("Content-Type")
		if !fragment.IsSupportedType(fragmentType) {
			c.JSON(http.StatusUnsupportedMediaType, gin.H{"message": "The specified file format is currently not supported!"})
			logger.Sugar.Infof("User tried to store fragment of type %s", fragmentType)
			return
		}
		fragment := fragment.Fragment{
			Id:      fragment_id,
			OwnerId: username, Created: time.Now(),
			Updated:      time.Now(),
			FragmentType: fragmentType,
			Size:         len(fileData)}
		fragment.SetData(fileData)
		logger.Sugar.Infof("File data being saved: %s", fileData)
		fragment.Save()

		scheme := "http://"
		if c.Request.TLS != nil {
			scheme = "https://"
		}

		c.Header("Location", scheme+c.Request.Host+fmt.Sprintf("/v1/fragment/%s", fragment.Id))
		c.JSON(http.StatusCreated, gin.H{"status": "ok", "message": "Fragment has successfully been updated",
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
		frag, err := fragment.GetFragment(hashing.HashString(username), fragment_id)
		logger.Sugar.Infof("Request to fetch fragments. User ID: %s. Fragment_id: %s", hashing.HashString(username), fragment_id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error"})
			logger.Sugar.Error("Failed to find user's fragments. Check if the username was hashed successfully")
			return
		}
		if frag == nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "Failed to find fragments with " +
				"the specified user id and fragment_id"})
			logger.Sugar.Error("Failed to find user's fragments. Check if the username was hashed successfully")
			return
		}
		logger.Sugar.Info("File type: ", frag.MimeType())
		var fileData []byte
		var mimeType string
		logger.Sugar.Info("Extension: ", ext)
		if ext == "" {
			fileData, err = frag.GetData()
			mimeType = frag.MimeType()
			if err != nil {
				logger.Sugar.Info("Failed to find the fragment")
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to find the fragment data"})
				return
			}
		} else {
			fileData, mimeType, err = frag.ConvertMimetype(ext)
			if err != nil {
				logger.Sugar.Info(err)
				c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to convert!"})
				return
			}
		}
		logger.Sugar.Info("Data in file: ", string(fileData))
		c.Header("Content-Length", strconv.Itoa(len(fileData)))
		c.Data(200, mimeType, fileData)
	})

	v1.GET("/fragment/:id/info", func(c *gin.Context) {
		id := c.Param("id")

		fragment, err := fragment.GetFragment(hashing.HashString(c.GetString("username")), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Server side error"})
		}
		if fragment == nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "Unable to find the specified fragment!"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok", "fragment": fragment})
	})

	v1.DELETE("/fragments/:id", func(c *gin.Context) {
		fragment_id := c.Param("id")
		if fragment_id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"id": "Please enter a valid ID!"})
			c.Next()
			return
		}
		username := c.GetString("username")
		fragment.DeleteFragment(hashing.HashString(username), fragment_id)
		c.JSON(http.StatusAccepted, gin.H{"message": "Fragment with ID" + fragment_id + " has been deleted!"})
	})

	return r
}

func main() {
	config.Config()
	// Create and assign logger instance to the global variable

	// UPLOADING TO DYNAMO DB
	// frag := fragment.Fragment{
	// 	Id:           "jashansfrag",
	// 	OwnerId:      "Jashan",
	// 	Created:      time.Now(),
	// 	Updated:      time.Now(),
	// 	FragmentType: "text/plain",
	// 	Size:         5,
	// }

	// frag.SetData([]byte("Hello"))
	// frag.Save()
	// newFrag, _ := fragment.GetFragment("324407508241184488", "0")
	// logger.Sugar.Info(newFrag)

	// UPLOADING TO S3
	// logger.Sugar.Info(fragment.GetS3Client().UploadFragmentDataToS3("jashan", "something", []byte("abcde")))
	// response, err := fragment.GetS3Client().GetFragmentDataFromS3("jashan", "something")
	// if err != nil {
	// 	logger.Sugar.Info(err)
	// }
	// logger.Sugar.Info(string(response))

	// Start server
	port := os.Getenv("PORT")
	r := getRouter()
	err := r.Run(port)
	if err != nil {
		logger.Sugar.Fatal("Failed to start the server")
	}
}
