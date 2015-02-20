package main

import (
	"os"
	"path/filepath"
	"time"
)

const (
	timeFormat = "2006-01-02 15:04:05"
	dateFormat = "2006-01-02"
)

func appPath() string {
	path, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return path
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
		logger.Fatalln(msg, err)
	}
}

func checkTimeFormat(timeStr string) bool {
	_, err := time.Parse(timeFormat, timeStr)

	if err != nil {
		_, err = time.Parse(dateFormat, timeStr)
	}

	return err == nil
}

func parseTime(timeStr string, clockToEnd bool) (time.Time, error) {
	t, err := time.ParseInLocation(timeFormat, timeStr, time.Local)

	if err != nil {
		t, err = time.ParseInLocation(dateFormat, timeStr, time.Local)
		if err == nil && clockToEnd {
			t = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, time.Local)
		}
	}

	return t, err
}

func indexOfString(arr []string, el string) int {
	for i, arrEl := range arr {
		if arrEl == el {
			return i
		}
	}

	return -1
}

func uniqStrings(arr []string) []string {
	if len(arr) < 2 {
		return arr
	}

	newArr := []string{}

	for _, el := range arr {
		if indexOfString(newArr, el) == -1 {
			newArr = append(newArr, el)
		}
	}

	return newArr
}
