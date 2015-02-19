package main

import (
	"testing"
)

func Test_absPathToFile(t *testing.T) {
	result := absPathToFile("/lorem/ipsum/dolor")
	if result != "/lorem/ipsum/dolor" {
		t.Error("Expected `/lorem/ipsum/dolor`, got ", result)
	}

	result = absPathToFile("./lorem/ipsum/dolor")
	if result != appPath()+"/lorem/ipsum/dolor" {
		t.Error("Expected `", appPath(), "/lorem/ipsum/dolor`, got ", result)
	}
}

func Test_checkTimeFormat(t *testing.T) {
	if !checkTimeFormat("2014-08-08") {
		t.Error("Expected `2014-08-08` to be valid format")
	}

	if !checkTimeFormat("2014-08-08 01:02:03") {
		t.Error("Expected `2014-08-08 01:02:03` to be valid format")
	}

	if checkTimeFormat("2014-08-08-90") {
		t.Error("Expected `2014-08-08-90` to be invalid format")
	}

	if checkTimeFormat("2014-08-08 01:02:03:04") {
		t.Error("Expected `2014-08-08 01:02:03:04` to be invalid format")
	}
}

func Test_parseTime(t *testing.T) {
	result, _ := parseTime("2014-09-08", false)
	if result.Day() != 8 || result.Month() != 9 || result.Year() != 2014 {
		t.Error("Expected `2014-09-08` to be parsed properly")
	}

	result, _ = parseTime("2014-09-08 11:12:13", false)
	if result.Hour() != 11 || result.Minute() != 12 || result.Second() != 13 {
		t.Error("Expected `2014-09-08 11:12:13` to be parsed properly")
	}
}

func Test_indexOfString(t *testing.T) {
	input := []string{"aaa", "bbb", "ccc"}

	result := indexOfString(input, "bbb")
	if result != 1 {
		t.Error("Expected 1, got ", result)
	}

	result = indexOfString(input, "ddd")
	if result != -1 {
		t.Error("Expected -1, got ", result)
	}
}

func Test_uniqStrings(t *testing.T) {
	input := []string{"fff", "fff2"}
	result := uniqStrings(input)
	if !(result[0] == input[0] && result[1] == input[1]) {
		t.Error("Expected [", input, "], got [", result, "]")
	}

	input = []string{"fff", "fff"}
	result = uniqStrings(input)
	if len(result) != 1 || result[0] != "fff" {
		t.Error("Expected [fff] got ", result)
	}
}
