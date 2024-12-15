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

type Metadata struct {
	UpdatedAt    int64            `json:"updatedAt,omitempty"`
	UpdatedAtMap map[string]int64 `json:"updatedAtMap,omitempty"`
}

type APIResponse struct {
	Success  bool        `json:"success"`
	Message  string      `json:"message,omitempty"`
	Data     interface{} `json:"data,omitempty"`
	Metadata *Metadata   `json:"metadata,omitempty"`
}

func setCommonHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}

func ErrorRes(w http.ResponseWriter, errMsg string, statusCode int) {
	w.WriteHeader(statusCode)
	setCommonHeaders(w)
	err := json.NewEncoder(w).Encode(APIResponse{Success: false, Message: errMsg})

	if err != nil {
		ErrorRes(w, ErrorMarshalling, http.StatusInternalServerError)
		return
	}
}

func SuccessResData(w http.ResponseWriter, data interface{}) {
	w.WriteHeader(http.StatusOK)
	setCommonHeaders(w)
	err := json.NewEncoder(w).Encode(APIResponse{Success: true, Data: data})

	if err != nil {
		ErrorRes(w, ErrorMarshalling, http.StatusInternalServerError)
		return
	}
}

func SuccessResDataWithMetadata(w http.ResponseWriter, data interface{}, m *Metadata) {
	w.WriteHeader(http.StatusOK)
	setCommonHeaders(w)
	err := json.NewEncoder(w).Encode(APIResponse{Success: true, Data: data, Metadata: m})
	if err != nil {
		ErrorRes(w, ErrorMarshalling, http.StatusInternalServerError)
		return
	}
}

func SuccessResMsg(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusOK)
	setCommonHeaders(w)
	err := json.NewEncoder(w).Encode(APIResponse{Success: true, Message: msg})
	if err != nil {
		ErrorRes(w, ErrorMarshalling, http.StatusInternalServerError)
		return
	}
}

func SuccessResMsgWithMetadata(w http.ResponseWriter, msg string, m *Metadata) {
	w.WriteHeader(http.StatusOK)
	setCommonHeaders(w)
	err := json.NewEncoder(w).Encode(APIResponse{Success: true, Message: msg, Metadata: m})
	if err != nil {
		ErrorRes(w, ErrorMarshalling, http.StatusInternalServerError)
		return
	}
}

func SuccessResMsgWithBody(w http.ResponseWriter, msg string, data interface{}) {
	w.WriteHeader(http.StatusOK)
	setCommonHeaders(w)
	err := json.NewEncoder(w).Encode(APIResponse{Success: true, Message: msg, Data: data})
	if err != nil {
		ErrorRes(w, ErrorMarshalling, http.StatusInternalServerError)
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
