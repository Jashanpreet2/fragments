package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	cognitoJwtVerify "github.com/jhosan7/cognito-jwt-verify"
)

func GetBody(bodyBytes []byte) map[string]interface{} {
	var jsonMap map[string]interface{}
	err := json.Unmarshal(bodyBytes, &jsonMap)
	if err != nil {
		log.Fatal(err)
	}
	return jsonMap
}

func GetUsername(c *gin.Context, local bool) (string, bool) {
	if local {
		if username, _, ok := c.Request.BasicAuth(); ok {
			return username, true
		} else {
			return "", false
		}
	} else {
		cognitoConfig := cognitoJwtVerify.Config{
			UserPoolId: os.Getenv("AWS_COGNITO_POOL_ID"),
			ClientId:   os.Getenv("AWS_COGNITO_CLIENT_ID"),
			TokenUse:   "id",
		}

		// sugar.Info("Authentication in process")
		token := c.GetHeader("Authorization")
		if !strings.HasPrefix(token, "Bearer ") {
			return "", false
		}

		token = token[len("Bearer "):]
		if token == "" {
			return "", false
		}

		verify, err := cognitoJwtVerify.Create(cognitoConfig)
		if err != nil {
			return "", false
		}

		payload, err := verify.Verify(token)
		if err != nil {
			return "", false
		}

		jsonData, err := json.Marshal(payload)
		if err != nil {
			return "", false
		}

		sugar.Info(jsonData)
		var v map[string]string
		json.Unmarshal(jsonData, &v)
		return v["cognito:username"], true
	}
}
