package main

import (
	"io"
	"log"
	"os"
)

var (
	logger  *log.Logger
	logFile os.File
)

func initLogger() {
	logFile, err := os.OpenFile(absPathToFile(config.Log.Path), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}

	writer := io.MultiWriter(os.Stdout, logFile)

	logger = log.New(writer, "Logbook: ", log.Ldate|log.Ltime)
}

func closeLogger() {
	logFile.Close()
}
