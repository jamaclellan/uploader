package http_responses

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ResponseHeader struct {
	Ok      bool   `json:"ok"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type JSONError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ErrorHolder interface {
	SetError(int, string)
}

type BaseResponse struct {
	ResponseHeader
	Results any `json:"results"`
}

func (h *ResponseHeader) SetError(code int, message string) {
	h.Ok = false
	h.Code = code
	h.Message = message
}

type HTTPError struct {
	status  int
	code    int
	message string
}

func (e HTTPError) Error() string {
	return fmt.Sprintf("%d: %s", e.code, e.message)
}

func Error(w http.ResponseWriter, target ErrorHolder, status, code int, message string) {
	target.SetError(code, message)
	Json(w, target, status)
}

func ErrorFromError(w http.ResponseWriter, target ErrorHolder, err error) {
	if httpError, ok := err.(*HTTPError); ok {
		Error(w, target, httpError.status, httpError.code, httpError.message)
		return
	}
	Error(w, target, http.StatusInternalServerError, -5000, "internal error")
}

func Json(w http.ResponseWriter, target any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	body, err := json.Marshal(target)
	if err != nil {
		// TODO: Handle this further?
		return
	}
	w.Write(body)
}
