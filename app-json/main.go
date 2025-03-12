package main

import (
	"log"
	"net/http"

	"github.com/rocketssan/toolkit"
)

type RequestPayload struct {
	Action  string `json:"action"`
	Message string `json:"message"`
}

type ResponsePayload struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status_code,omitempty"`
}

func main() {
	mux := routes()

	log.Println("starting server")

	err := http.ListenAndServe(":8081", mux)
	if err != nil {
		log.Fatal(err)
	}
}

func routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("."))))
	mux.HandleFunc("/receive-post", receivePost)
	mux.HandleFunc("/remote-service", remoteService)
	mux.HandleFunc("/simulated-service", simulatedService)

	return mux
}

func receivePost(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload
	var tool toolkit.Tools

	err := tool.ReadJSON(w, r, &requestPayload)
	if err != nil {
		tool.ErrorJSON(w, err)
		return
	}

	responsePayload := ResponsePayload{
		Message: "hit the handler okay, and sending response",
	}

	err = tool.WriteJSON(w, http.StatusOK, responsePayload)
	if err != nil {
		tool.ErrorJSON(w, err)
		return
	}
}

func remoteService(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload
	var tool toolkit.Tools

	err := tool.ReadJSON(w, r, &requestPayload)
	if err != nil {
		tool.ErrorJSON(w, err)
		return
	}

	_, statusCode, err := tool.PushJSONToRemote("http://localhost:8081/simulated-service", http.MethodPost, requestPayload)
	if err != nil {
		tool.ErrorJSON(w, err)
		return
	}

	responsePayload := ResponsePayload{
		Message:    "hit the handler okay, and sending response",
		StatusCode: statusCode,
	}

	err = tool.WriteJSON(w, http.StatusOK, responsePayload)
	if err != nil {
		tool.ErrorJSON(w, err)
		return
	}
}

func simulatedService(w http.ResponseWriter, _ *http.Request) {
	payload := ResponsePayload{
		Message: "OK",
	}

	var tool toolkit.Tools

	_ = tool.WriteJSON(w, http.StatusOK, payload)
}
