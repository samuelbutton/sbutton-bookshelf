package main

import (
	"encoding/json"
	"net/http"
)

// utils.go is used to build json messages and return a json response

// Message builds a json message to send
func Message(status bool, message string) map[string]interface{} {
	return map[string]interface{}{"status": status, "message": message}
}

// Respond writes a json message
func Respond(w http.ResponseWriter, data map[string]interface{}) {
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
