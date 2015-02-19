package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type RequestHandler func(http.ResponseWriter, *http.Request)

// Server tools ====================================================================================

func startServer() {
	router := mux.NewRouter()

	router.HandleFunc("/{application}/put", basicAuth(createLogHandler)).
		Methods("POST")
	router.HandleFunc("/{application}/get", basicAuth(getLogsHandler)).
		Methods("GET")

	bindAddress := config.Server.Address + ":" + config.Server.Port

	logger.Printf("Starting server on %s\n", bindAddress)

	err := http.ListenAndServe(bindAddress, router)
	if err != nil {
		logger.Fatalf("Can't start server: %v", err)
	}
}

func basicAuth(handler RequestHandler) RequestHandler {
	return func(rw http.ResponseWriter, req *http.Request) {
		if len(req.Header["Authorization"]) == 0 {
			serverError(rw, errors.New("Authorization required"), 401)
			return
		}

		auth := strings.SplitN(req.Header["Authorization"][0], " ", 2)

		if len(auth) != 2 || auth[0] != "Basic" {
			serverError(rw, errors.New("Bad syntax"), 400)
			return
		}

		payload, _ := base64.StdEncoding.DecodeString(auth[1])
		pair := strings.SplitN(string(payload), ":", 2)

		if len(pair) != 2 || pair[0] != config.Auth.User || pair[1] != config.Auth.Password {
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
