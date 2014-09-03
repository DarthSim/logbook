package main

import (
	"flag"
	"fmt"
	"os"

	"code.google.com/p/gcfg"
)

type Config struct {
	Server struct {
		Address string
		Port    string
	}
	Database struct {
		Path        string
		LockTimeout int64
		RetryDelay  int64
	}
	Log struct {
		Path        string
		LogDatabase bool
	}
}

var config Config

func init() {
	configfile := flag.String("config", "../logbook.conf", "path to configuration file")

	err := gcfg.ReadFileInto(&config, absPathToFile(*configfile))
	if err != nil {
		fmt.Printf("Error opening config file: %v", err)
		os.Exit(1)
	}
}
