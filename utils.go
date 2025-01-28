package main

import (
	"encoding/json"
)

func GetBody(bodyBytes []byte) map[string]interface{} {
	var jsonMap map[string]interface{}
	json.Unmarshal(bodyBytes, &jsonMap)
	return jsonMap
}
