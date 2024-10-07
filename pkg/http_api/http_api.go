package http_api

import (
	"encoding/json"
	"net/http"
)

const (
	ErrorInvalidRequest        = "Invalid request"
	ErrorRouteNotFound         = "Route not found"
	ErrorMethodNotAllowed      = "Method not allowed"
	ErrorEmptyLambdaEvent      = "Empty lambda event"
	ErrorCouldNotMarshalItem   = "Could not marshal item"
	ErrorCouldNotUnMarshalItem = "Could not  unmarshal item"
)

type ErrorBody struct {
	ErrorMsg *string `json:"error,omitempty"`
}

type RespBody struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func SuccessResData(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(RespBody{Success: true, Data: data})
}

func SuccessResMsg(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(RespBody{Success: true, Message: msg})
}

func SuccessResMsgWithBody(w http.ResponseWriter, msg string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(RespBody{Success: true, Message: msg, Data: data})
}
