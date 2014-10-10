package main

import (
	"testing"
)

func Test_checkTimeFormat(t *testing.T) {
	if !checkTimeFormat("2014-08-08") {
		t.Error("failed")
	}

	if checkTimeFormat("2014-08-08-90") {
		t.Error("failed")
	}
}

func Test_parseTime(t *testing.T) {
	result, _ := parseTime("2014-09-08", false)
	if result.Day() != 8 {
		t.Error("failed")
	}

	if result.Month() != 9 {
		t.Error("failed")
	}
}

func Test_uniqStrings_Good(t *testing.T) {
	input := []string{"fff", "fff2"}
	result := uniqStrings(input)
	if !(result[0] == input[0] && result[1] == input[1]) {
		t.Error("failed")
	}
}

func Test_uniqStrings_Fail(t *testing.T) {
	input := []string{"fff", "fff"}
	result := uniqStrings(input)
	if len(result) != 1 {
		t.Error("failed")
	}
}
