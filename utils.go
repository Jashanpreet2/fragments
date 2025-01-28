package main

import (
	"encoding/json"
	"log"
)

func GetBody(bodyBytes []byte) map[string]interface{} {
	var jsonMap map[string]interface{}
	err := json.Unmarshal(bodyBytes, &jsonMap)
	if err != nil {
		log.Print("Failed to retrieve json values")
	}
	return jsonMap
}
