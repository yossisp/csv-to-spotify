package server

import (
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/yossisp/csv-to-spotify/pkg/utils"

	"github.com/yossisp/csv-to-spotify/pkg/config"

	"github.com/yossisp/csv-to-spotify/pkg/runner"
	"github.com/yossisp/csv-to-spotify/pkg/websocket"

	"github.com/yossisp/csv-to-spotify/pkg/db"
	"github.com/rs/cors"
)

var (
	conf   = config.NewConfig()
	logger = utils.NewLogger("server")
)

type serverErrorType int

const (
	CSVError serverErrorType = iota + 1
)

type serverError struct {
	ErrorMessage serverErrorType `json:"errorMessage"`
}

func sendError(w http.ResponseWriter, errorType serverErrorType) {
	const funcName = "sendError"
	result := serverError{ErrorMessage: errorType}
	resultJSON, err := json.Marshal(result)
	if err != nil {
		logger("%s: json.Marshal: %v", funcName, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	http.Error(w, string(resultJSON), http.StatusBadRequest)
}

// InitServer starts the server
func InitServer() {
	const funcName = "InitServer"
	cors := cors.New(cors.Options{
		AllowOriginFunc: func(origin string) bool {
			return strings.Contains(conf.AllowedOrigins, origin)
		},
	})
	mux := http.NewServeMux()
	userHandler := func(w http.ResponseWriter, req *http.Request) {
		const funcName = "userHandler"
		logger("%s: new request from: %s", funcName, req.URL.String())
		user := db.SpotifyUser{}
		err := json.NewDecoder(req.Body).Decode(&user)
		if err != nil {
			logger("%s: json.NewDecoder: %v", funcName, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		db.InsertSpotifyUser(user)
	}

	csvHandler := func(w http.ResponseWriter, req *http.Request) {
		const funcName = "csvHandler"
		payload := runner.CSVPayload{}
		err := json.NewDecoder(req.Body).Decode(&payload)
		if err != nil {
			logger("%s: json.NewDecoder: %v", funcName, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		dbUser := db.FindSpotifyUser(*payload.UserID)
		if payload.UserID == nil || payload.CSVFile == nil || dbUser == nil {
			errMsg := "missing/bad userId or missing csvFile"
			logger("%s: %s", funcName, errMsg)
			sendError(w, CSVError)
			return
		}
		fileName := *payload.FileName
		// client logic enforces that the file is csv
		*payload.FileName = strings.TrimSuffix(fileName, filepath.Ext(fileName))
		runner := runner.NewRunner(payload, dbUser)
		go runner.Run()
	}
	mux.HandleFunc("/csv", csvHandler)
	mux.HandleFunc("/user", userHandler)
	mux.HandleFunc("/websocket", websocket.WSConnectionHandler)
	handler := cors.Handler(mux)
	logger("%s: Listing for requests at port %s", funcName, conf.Port)
	log.Fatal(http.ListenAndServe(":"+conf.Port, handler))
}
