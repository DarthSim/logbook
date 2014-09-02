package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"code.google.com/p/gcfg"
)

type Config struct {
	Server struct {
		Address string
		Port    string
	}
	Database struct {
		Path string
	}
	Log struct {
		Path        string
		LogDatabase bool
	}
}

var config Config

func init() {
	configfile := flag.String("config", defaultConfigPath(), "path to configuration file")

	err := gcfg.ReadFileInto(&config, *configfile)
	if err != nil {
		fmt.Printf("Error opening config file: %v", err)
		os.Exit(1)
	}
}

func defaultConfigPath() string {
	app_path, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return filepath.Join(app_path, "../logbook.conf")
}
