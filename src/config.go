package main

import (
	"flag"
	"fmt"
	"os"

	"code.google.com/p/gcfg"
)

type Config struct {
	Auth struct {
		User     string
		Password string
	}
	Server struct {
		Address string
		Port    string
	}
	Database struct {
		Path string
	}
	Log struct {
		Path string
	}
	Pagination struct {
		PerPage int
	}
}

var config Config

func prepareConfig() {
	configfile := flag.String(
		"config",
		"../logbook.conf",
		"path to configuration file",
	)

	flag.Parse()

	if err := gcfg.ReadFileInto(&config, absPathToFile(*configfile)); err != nil {
		fmt.Printf("Error opening config file: %v", err)
		os.Exit(1)
	}
}
