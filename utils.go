package main

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

const dateFormat = "2006-01-02"

var timeFormats = [...]string{
	"2006-01-02T15:04:05",
	"2006-01-02T15:04:05-07:00",
	"2006-01-02T15:04:05.000",
	"2006-01-02T15:04:05.000-07:00",
}

func appPath() (path string) {
	path, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	return
}

func absPathToFile(path string) string {
	if filepath.IsAbs(path) {
		return path
	} else {
		return filepath.Join(appPath(), path)
	}
}

func checkErr(err error, msg string) {
	if err != nil {
		log.Fatalf("%s (%v)", msg, err)
	}
}

func checkTimeFormat(timeStr string) bool {
	for _, format := range timeFormats {
		_, err := time.Parse(format, timeStr)
		if err == nil {
			return true
		}
	}

	return false
}

func checkDateTimeFormat(timeStr string) bool {
	if checkTimeFormat(timeStr) {
		return true
	}

	_, err := time.Parse(dateFormat, timeStr)
	return err == nil
}

func parseTime(timeStr string) (t time.Time, err error) {
	for _, format := range timeFormats {
		t, err = time.ParseInLocation(format, timeStr, time.Local)
		if err == nil {
			t = t.Local()
			return
		}
	}

	return
}

func parseDateTime(timeStr string, clockToEnd bool) (t time.Time, err error) {
	t, err = parseTime(timeStr)

	if err != nil {
		t, err = time.ParseInLocation(dateFormat, timeStr, time.Local)
		if err == nil && clockToEnd {
			t = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, time.Local)
		}
	}

	return
}

func uniqStrings(arr []string) []string {
	if len(arr) < 2 {
		return arr
	}

	m := make(map[string]struct{})
	for _, el := range arr {
		m[el] = struct{}{}
	}

	newArr := make([]string, len(m))
	i := 0
	for el := range m {
		newArr[i] = el
		i++
	}

	return newArr
}

func stringsContain(arr1, arr2 []string) bool {
	smap := make(map[string]struct{})

	for _, str := range arr1 {
		smap[str] = struct{}{}
	}
	for _, str := range arr2 {
		if _, ok := smap[str]; !ok {
			return false
		}
	}

	return true
}

func panicOnErr(err error) {
	if err != nil {
		log.Panic(err)
	}
}
