package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_absPathToFile(t *testing.T) {
	assert.Equal(t, "/lorem/ipsum/dolor",
		absPathToFile("/lorem/ipsum/dolor"))

	assert.Equal(t, appPath()+"/lorem/ipsum/dolor",
		absPathToFile("./lorem/ipsum/dolor"))
}

func Test_checkTimeFormat(t *testing.T) {
	assert.True(t, checkTimeFormat("2014-08-08"),
		"2014-08-08 should be responded as valid date")

	assert.True(t, checkTimeFormat("2014-08-08 01:02:03"),
		"2014-08-08 01:02:03 should be responded as valid datetime")

	assert.False(t, checkTimeFormat("2014-08-08-90"),
		"2014-08-08-90 should be responded as valid date")

	assert.False(t, checkTimeFormat("2014-08-08 01:02:03:04"),
		"2014-08-08 01:02:03:04 should be responded as valid datetime")
}

func Test_parseTime(t *testing.T) {
	result, _ := parseTime("2014-09-08", false)
	assert.EqualValues(t, time.Date(2014, 9, 8, 0, 0, 0, 0, time.Local),
		result)

	result, _ = parseTime("2014-09-08", true)
	assert.EqualValues(t, time.Date(2014, 9, 8, 23, 59, 59, 999999999, time.Local),
		result)

	result, _ = parseTime("2014-09-08 11:12:13", false)
	assert.EqualValues(t, time.Date(2014, 9, 8, 11, 12, 13, 0, time.Local),
		result)
}

func Test_indexOfString(t *testing.T) {
	input := []string{"aaa", "bbb", "ccc"}

	assert.Equal(t, 1, indexOfString(input, "bbb"))
	assert.Equal(t, -1, indexOfString(input, "ddd"))
}

func Test_uniqStrings(t *testing.T) {
	input := []string{"fff", "fff2"}
	assert.Equal(t, input, uniqStrings(input))

	input = []string{"fff", "fff"}
	assert.Equal(t, []string{"fff"}, uniqStrings(input))
}
