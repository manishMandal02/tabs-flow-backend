package http_api

import (
	"encoding/json"
	"net/http"
)

const (
	ErrorInvalidRequest   = "Invalid request"
	ErrorRouteNotFound    = "Route not found"
	ErrorMethodNotAllowed = "Method not allowed"
	ErrorEmptyLambdaEvent = "Empty lambda event"
	ErrorMarshalling      = "Error marshaling"
	ErrorUnMarshalling    = "Error un_marshaling "
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
	err := json.NewEncoder(w).Encode(RespBody{Success: true, Data: data})

	if err != nil {
		http.Error(w, ErrorMarshalling, http.StatusInternalServerError)
		return
	}
}

func SuccessResMsg(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(RespBody{Success: true, Message: msg})
	if err != nil {
		http.Error(w, ErrorMarshalling, http.StatusInternalServerError)
		return
	}
}

func SuccessResMsgWithBody(w http.ResponseWriter, msg string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(RespBody{Success: true, Message: msg, Data: data})
	if err != nil {
		http.Error(w, ErrorMarshalling, http.StatusInternalServerError)
		return
	}
}

type responseWriterWritten struct {
	http.ResponseWriter
	Written bool
}

func (w *responseWriterWritten) WriteHeader(status int) {
	w.Written = true
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseWriterWritten) Write(b []byte) (int, error) {
	w.Written = true
	return w.ResponseWriter.Write(b)
}

func (w *responseWriterWritten) HasWritten() bool {
	return w.Written

}
