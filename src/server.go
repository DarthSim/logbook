package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

// Server tools ====================================================================================

func startServer() {
	router := mux.NewRouter()

	router.HandleFunc("/{application}/put", createLogHandler).Methods("POST")
	router.HandleFunc("/{application}/get", getLogsHandler).Methods("GET")

	bindAddress := config.Server.Address + ":" + config.Server.Port

	logger.Printf("Starting server on %s\n", bindAddress)

	err := http.ListenAndServe(bindAddress, router)
	if err != nil {
		logger.Fatalf("Can't start server: %v", err)
	}
}

func requestVars(req *http.Request) map[string]string {
	return mux.Vars(req)
}

func serverError(rw http.ResponseWriter, err error, status int) {
	logger.Printf("Server error: %v", err)

	response, _ := json.Marshal(ErrorResponse{
		Error: err.Error(),
	})

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(status)
	rw.Write(response)
}

func serverResponse(rw http.ResponseWriter, response []byte) {
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.Write(response)
}

func safeParseForm(r *http.Request) error {
	err := r.ParseMultipartForm(32 << 10)

	if err == http.ErrNotMultipart {
		err = r.ParseForm()
	}

	return err
}

// end of Server tools
