package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type RequestHandler func(http.ResponseWriter, *http.Request)

// Server tools ====================================================================================

func startServer() {
	bindAddress := config.Server.Address + ":" + config.Server.Port

	logger.Printf("Starting server on %s\n", bindAddress)

	if err := http.ListenAndServe(bindAddress, setupRouter()); err != nil {
		logger.Fatalf("Can't start server: %v", err)
	}
}

func setupRouter() (router *mux.Router) {
	router = mux.NewRouter()

	router.HandleFunc("/{application}/put", basicAuth(createLogHandler)).
		Methods("POST")
	router.HandleFunc("/{application}/get", basicAuth(getLogsHandler)).
		Methods("GET")

	return
}

func basicAuth(handler RequestHandler) RequestHandler {
	return func(rw http.ResponseWriter, req *http.Request) {
		username, password, ok := req.BasicAuth()

		if !ok || username != config.Auth.User || password != config.Auth.Password {
			serverError(rw, errors.New("Authorization failed"), 401)
			return
		}

		handler(rw, req)
	}
}

func requestVars(req *http.Request) map[string]string {
	return mux.Vars(req)
}

func serverError(rw http.ResponseWriter, err error, status int) {
	var msg string

	if status > 500 {
		logger.Printf("Server error: %v", err)
		msg = "Internal server error"
	} else {
		msg = err.Error()
	}

	response, _ := json.Marshal(ErrorResponse{msg})

	rw.WriteHeader(status)
	serverResponse(rw, response)
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
