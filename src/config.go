package main

import (
	"flag"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
	"io/ioutil"
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
	Pagination struct {
		PerPage int
	}
}

var config Config

func prepareConfig() {
	configfile := flag.String(
		"config",
		"../logbook.yml",
		"path to configuration file",
	)

	flag.Parse()

	var (
		confData []byte
		err      error
	)

	if confData, err = ioutil.ReadFile(absPathToFile(*configfile)); err != nil {
		fmt.Printf("Error opening config file: %v", err)
		os.Exit(1)
	}

	if err = yaml.Unmarshal(confData, &config); err != nil {
		fmt.Printf("Invalid config file format")
		os.Exit(1)
	}
}
