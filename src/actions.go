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

func checkCommonParams(lvl string, tags []string) error {
	if len(lvl) != 1 || lvl < "0" || lvl > "5" {
		return errors.New("Level should be a number between 0 and 5")
	}

	for _, tag := range tags {
		if tag == "" {
			return errors.New("Tags contain an empty string")
		}
	}

	return nil
}

// Action: Create log ==========================================================

func checkCreateLogParams(msg string, lvl string, tags []string, createdAt string) error {
	if msg == "" {
		return errors.New("Message should be defined")
	}

	if err := checkCommonParams(lvl, tags); err != nil {
		return err
	}

	if len(createdAt) > 0 && !checkTimeFormat(createdAt) {
		return errors.New("Created at has invalid format")
	}

	return nil
}

func createLogHandler(c *gin.Context) {
	application := c.Param("application")
	message := c.PostForm("message")
	levelStr := c.PostForm("level")
	tags := uniqStrings(extractTags(c.PostForm("tags")))
	createdAtStr := c.PostForm("created_at")

	if err := checkCreateLogParams(message, levelStr, tags, createdAtStr); err != nil {
		c.JSON(422, ErrorResponse{err.Error()})
		return
	}

	logRecord := LogRecord{
		Message: message,
		Tags:    tags,
	}

	logRecord.Level, _ = strconv.Atoi(levelStr)

	if len(createdAtStr) > 0 {
		logRecord.CreatedAt, _ = parseTime(createdAtStr)
	}

	panicOnErr(storage.SaveLogRecord(application, &logRecord))

	c.JSON(200, logRecord)
}

// end of Action: Create log

// Action: Get logs ============================================================

func checkGetLogParams(lvl string, tags []string, startTime string, endTime string, page string) error {
	if err := checkCommonParams(lvl, tags); err != nil {
		return err
	}

	if !checkDateTimeFormat(startTime) {
		return errors.New("Start time has invalid format")
	}

	if !checkDateTimeFormat(endTime) {
		return errors.New("End time has invalid format")
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
	tags := uniqStrings(extractTags(c.Query("tags")))
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
	startTime, _ := parseDateTime(startTimeStr, false)
	endTime, _ := parseDateTime(endTimeStr, true)

	logRecords, err := storage.LoadLogRecords(
		application, level, tags, startTime, endTime, page,
	)
	panicOnErr(err)

	c.JSON(200, logRecords)
}

// end of Action: Get logs

// Action: App stats ===========================================================

func appStatsHandler(c *gin.Context) {
	stats, err := storage.appStats(c.Param("application"))
	panicOnErr(err)
	c.String(200, stats)
}

// end of Action: App stats
