package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type ConfigMap map[string]string

type Config struct {
	Username string
	Password string

	Address string
	Port    int

	DBPath        string
	DBCompression int

	RecordsPerPage int
}

var config Config

var compressionTypes map[string]int = map[string]int{
	"NoCompression": 0,
	"Snappy":        1,
	"Zlib":          2,
	"BZip2":         3,
	"LZ4":           4,
	"LZ4HC":         5,
}

func (c ConfigMap) Load(filename string) {
	file, err := os.Open(absPathToFile(filename))
	checkErr(err, "Error opening config file")
	defer file.Close()

	re, _ := regexp.Compile("\\s+")
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if len(line) == 0 || line[0] == '#' || line[0] == ';' {
			continue
		}

		param := re.Split(line, 2)

		if len(param) == 1 {
			continue
		}

		c[param[0]] = param[1]
	}
}

func (c ConfigMap) GetStr(name, def string) string {
	if val, ok := c[name]; ok {
		return val
	}
	return def
}

func (c ConfigMap) GetInt(name string, def int) int {
	if vals, ok := c[name]; ok {
		vali, err := strconv.Atoi(vals)
		checkErr(err, "Error reading config file")
		return vali
	}
	return def
}

func prepareConfig() {
	cpath := flag.String(
		"config", "../logbook.conf", "path to configuration file",
	)
	flag.Parse()

	cmap := make(ConfigMap)
	cmap.Load(*cpath)

	config.Username = cmap["username"]
	config.Password = cmap["password"]

	config.Address = cmap.GetStr("bind", "0.0.0.0")
	config.Port = cmap.GetInt("port", 11610)

	config.DBPath = absPathToFile(cmap.GetStr("db_path", "../db"))

	if comp, ok := compressionTypes[cmap.GetStr("db_compression", "Snappy")]; ok {
		config.DBCompression = comp
	} else {
		log.Fatalln("Invalid db_compression value")
	}

	config.RecordsPerPage = cmap.GetInt("records_per_page", 100)
}
