package server

import (
	"encoding/json"
)

type RequestMessage string
type ResponseMessage string

const MONVAN_EOF = "d2a81ac4c97aa94e0c3474c619cc85508a19544f388c9c1c0d26773a2b19eb09"

const (
	// Request types
	AUTHENTICATE RequestMessage = "authenticate"
	QUERY        RequestMessage = "query"
	CLOSE_CONN   RequestMessage = "close_conn"
)

const (
	// Response types
	ERROR   ResponseMessage = "error"
	SUCCESS ResponseMessage = "success"
)

type UserMessage struct {
	RequestType RequestMessage `json:"type"`
	Message     interface{}    `json:"message"`
}

type ErrorMessage struct {
	Message string `json:"message"`
}

type SuccessMessage struct {
	ResultJson string `json:"result"`
	Duration   int    `json:"duration"`
}

type AuthenticateMessage struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
}

type QueryMessage string

func BuildFailureMessage(message string) []byte {
	obj := map[string]interface{}{
		"type":    "error",
		"message": message,
	}

	res, err := json.Marshal(obj)
	if err != nil {
		return []byte("{\"type\":\"error\",\"message\":\"Failed to marshal error message\"}" + MONVAN_EOF)
	}
	if len(res) == 0 {
		return []byte("{\"type\":\"error\",\"message\":\"Empty error message\"}" + MONVAN_EOF)
	}
	if res[len(res)-1] != '}' {
		return []byte("{\"type\":\"error\",\"message\":\"Invalid error message format\"}" + MONVAN_EOF)
	}

	return append(res, MONVAN_EOF...)
}

func BuildSuccessMessage(result any, duration int) []byte {
	obj := map[string]interface{}{
		"type":     "success",
		"result":   result,
		"duration": duration,
	}
	res, err := json.Marshal(obj)
	if err != nil {
		return []byte("{\"type\":\"error\",\"message\":\"Failed to marshal result\"}" + MONVAN_EOF)
	}
	if len(res) == 0 {
		return []byte("{\"type\":\"error\",\"message\":\"Empty result\"}" + MONVAN_EOF)
	}
	if res[len(res)-1] != '}' {
		return []byte("{\"type\":\"error\",\"message\":\"Invalid result format\"}" + MONVAN_EOF)
	}
	return append(res, MONVAN_EOF...)
}
