package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Response formats ============================================================

type LogRecordResponse struct {
	Id          int64      `json:"id"`
	Application string     `json:"application"`
	Level       int        `json:"level"`
	CreatedAt   *time.Time `json:"created_at"`
	Message     string     `json:"message"`
	Tags        []string   `json:"tags"`
}

type LogRecordsResponse []LogRecordResponse

// end of Response formats

// Helpers =====================================================================
func extractTagNames(form url.Values) []string {
	if len(form["tags[]"]) > 0 {
		return form["tags[]"]
	} else if len(form["tags"]) > 0 {
		return strings.Split(form.Get("tags"), ",")
	}

	return make([]string, 0)
}

// end of Helpers

// Action: Create log ==========================================================

func buildCreateLogResponse(logRecord *LogRecord) ([]byte, error) {
	response := LogRecordResponse{
		Id:          logRecord.Id,
		Application: logRecord.Application.Name,
		Level:       logRecord.Level,
		CreatedAt:   logRecord.CreatedAt,
		Message:     logRecord.Message,
		Tags:        make([]string, len(logRecord.Tags)),
	}

	for i, tag := range logRecord.Tags {
		response.Tags[i] = tag.Name
	}

	return json.Marshal(response)
}

func checkCreateLogParams(msg string, lvl string, tagNames []string) error {
	if msg == "" {
		return errors.New("Message should be defined")
	}

	if len(lvl) != 1 || lvl < "1" || lvl > "5" {
		return errors.New("Level should be a number between 1 and 5")
	}

	for _, tagName := range tagNames {
		if tagName == "" {
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

	appName := vars["application"]
	message := req.Form.Get("message")
	levelStr := req.Form.Get("level")
	tagNames := extractTagNames(req.Form)

	if err = checkCreateLogParams(message, levelStr, tagNames); err != nil {
		serverError(rw, err, 422)
		return
	}

	level, _ := strconv.Atoi(levelStr)

	logRecord, err := createLogRecord(appName, message, level, tagNames)
	if err != nil {
		serverError(rw, err, 500)
		return
	}

	response, err := buildCreateLogResponse(&logRecord)
	if err != nil {
		serverError(rw, err, 500)
		return
	}

	serverResponse(rw, response)
}

// end of Action: Create log

// Action: Get logs ============================================================

func buildGetLogsResponse(logRecords *[]LogRecord) ([]byte, error) {
	response := make([]LogRecordResponse, len(*logRecords))

	for i, logRecord := range *logRecords {
		response[i] = LogRecordResponse{
			Id:          logRecord.Id,
			Application: logRecord.Application.Name,
			Level:       logRecord.Level,
			CreatedAt:   logRecord.CreatedAt,
			Message:     logRecord.Message,
			Tags:        make([]string, len(logRecord.Tags)),
		}
		for n, tag := range logRecord.Tags {
			response[i].Tags[n] = tag.Name
		}
	}

	return json.Marshal(response)
}

func checkGetLogParams(lvl string, tagNames []string, startTime string, endTime string, page string) error {
	if len(lvl) != 1 || lvl < "1" || lvl > "5" {
		return errors.New("Level should be a number between 1 and 5")
	}

	for _, tagName := range tagNames {
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

	appName := vars["application"]
	levelStr := req.Form.Get("level")
	startTimeStr := req.Form.Get("start_time")
	endTimeStr := req.Form.Get("end_time")
	tagNames := extractTagNames(req.Form)
	pageStr := req.Form.Get("page")

	if pageStr == "" {
		pageStr = "1"
	}

	err = checkGetLogParams(levelStr, tagNames, startTimeStr, endTimeStr, pageStr)
	if err != nil {
		serverError(rw, err, 422)
		return
	}

	level, _ := strconv.Atoi(levelStr)
	page, err := strconv.Atoi(pageStr)
	startTime, _ := parseTime(startTimeStr, false)
	endTime, _ := parseTime(endTimeStr, true)

	logRecords, err := findLogRecords(appName, level, tagNames, &startTime, &endTime, page)
	if err != nil {
		serverError(rw, err, 500)
		return
	}

	response, err := buildGetLogsResponse(&logRecords)
	if err != nil {
		serverError(rw, err, 500)
		return
	}

	serverResponse(rw, response)
}

// end of Action: Get logs
