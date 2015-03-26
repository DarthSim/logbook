package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// Helpers =====================================================================
func extractTags(form url.Values) []string {
	if len(form["tags[]"]) > 0 {
		return form["tags[]"]
	} else if len(form["tags"]) > 0 {
		return strings.Split(form.Get("tags"), ",")
	}

	return make([]string, 0)
}

// end of Helpers

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

func createLogHandler(rw http.ResponseWriter, req *http.Request) {
	var err error

	vars := requestVars(req)

	if err = safeParseForm(req); err != nil {
		serverError(rw, err, 500)
		return
	}

	application := vars["application"]
	message := req.Form.Get("message")
	levelStr := req.Form.Get("level")
	tags := extractTags(req.Form)

	if err = checkCreateLogParams(message, levelStr, tags); err != nil {
		serverError(rw, err, 422)
		return
	}

	level, _ := strconv.Atoi(levelStr)
	logRecord := LogRecord{
		Message: message,
		Level:   level,
		Tags:    tags,
	}

	if err = saveLogRecord(application, &logRecord); err != nil {
		serverError(rw, err, 500)
		return
	}

	response, err := json.Marshal(logRecord)
	if err != nil {
		serverError(rw, err, 500)
		return
	}

	serverResponse(rw, response)
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

func getLogsHandler(rw http.ResponseWriter, req *http.Request) {
	var err error

	vars := requestVars(req)

	if err = safeParseForm(req); err != nil {
		serverError(rw, err, 500)
		return
	}

	application := vars["application"]
	levelStr := req.Form.Get("level")
	startTimeStr := req.Form.Get("start_time")
	endTimeStr := req.Form.Get("end_time")
	tags := extractTags(req.Form)
	pageStr := req.Form.Get("page")

	if pageStr == "" {
		pageStr = "1"
	}

	err = checkGetLogParams(levelStr, tags, startTimeStr, endTimeStr, pageStr)
	if err != nil {
		serverError(rw, err, 422)
		return
	}

	level, _ := strconv.Atoi(levelStr)
	page, err := strconv.Atoi(pageStr)
	startTime, _ := parseTime(startTimeStr, false)
	endTime, _ := parseTime(endTimeStr, true)

	logRecords, err := loadLogRecords(application, level, tags, startTime, endTime, page)
	if err != nil {
		serverError(rw, err, 500)
		return
	}

	response, err := json.Marshal(&logRecords)
	if err != nil {
		serverError(rw, err, 500)
		return
	}

	serverResponse(rw, response)
}

// end of Action: Get logs
