package main

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func extractTags(tags string) []string {
	if len(tags) > 0 {
		return strings.Split(tags, ",")
	} else {
		return []string{}
	}
}

// Action: Create log ==========================================================

func checkCreateLogParams(msg string, lvl string, tags []string) error {
	if msg == "" {
		return errors.New("Message should be defined")
	}

	if len(lvl) != 1 || lvl < "1" || lvl > "5" {
		return errors.New("Level should be a number between 1 and 5")
	}

	for _, tag := range tags {
		if tag == "" {
			return errors.New("Tags contain an empty string")
		}
	}

	return nil
}

func createLogHandler(c *gin.Context) {
	application := c.Param("application")
	message := c.PostForm("message")
	levelStr := c.PostForm("level")
	tags := extractTags(c.PostForm("tags"))

	if err := checkCreateLogParams(message, levelStr, tags); err != nil {
		c.JSON(422, ErrorResponse{err.Error()})
		return
	}

	level, _ := strconv.Atoi(levelStr)
	logRecord := LogRecord{
		Message: message,
		Level:   level,
		Tags:    tags,
	}

	panicOnErr(saveLogRecord(application, &logRecord))

	c.JSON(200, logRecord)
}

// end of Action: Create log

// Action: Get logs ============================================================

func checkGetLogParams(lvl string, tags []string, startTime string, endTime string, page string) error {
	if len(lvl) != 1 || lvl < "1" || lvl > "5" {
		return errors.New("Level should be a number between 1 and 5")
	}

	for _, tagName := range tags {
		if tagName == "" {
			return errors.New("Tags contain an empty string")
		}
	}

	if !checkTimeFormat(startTime) {
		return errors.New("Start time should be YYYY-MM-DD or YYYY-MM-DD hh-mm-ss")
	}

	if !checkTimeFormat(endTime) {
		return errors.New("End time should be YYYY-MM-DD or YYYY-MM-DD hh-mm-ss")
	}

	if correct, _ := regexp.MatchString("\\A\\d+\\z", page); !correct {
		return errors.New("Page should be greater or equal to 1")
	}

	return nil
}

func getLogsHandler(c *gin.Context) {
	var err error

	application := c.Param("application")
	levelStr := c.Query("level")
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")
	tags := extractTags(c.Query("tags"))
	pageStr := c.Query("page")

	if pageStr == "" {
		pageStr = "1"
	}

	err = checkGetLogParams(levelStr, tags, startTimeStr, endTimeStr, pageStr)
	if err != nil {
		c.JSON(422, ErrorResponse{err.Error()})
		return
	}

	level, _ := strconv.Atoi(levelStr)
	page, err := strconv.Atoi(pageStr)
	startTime, _ := parseTime(startTimeStr, false)
	endTime, _ := parseTime(endTimeStr, true)

	logRecords, err := loadLogRecords(
		application, level, tags, startTime, endTime, page,
	)
	panicOnErr(err)

	c.JSON(200, logRecords)
}

// end of Action: Get logs
