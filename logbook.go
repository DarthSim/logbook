package main

import (
	"os"
	"os/signal"
)

func main() {
	prepareConfig()

	OpenStorage()
	defer storage.Close()

	startServer()

	interrupted := make(chan os.Signal)
	signal.Notify(interrupted, os.Interrupt, os.Kill)
	<-interrupted
}
