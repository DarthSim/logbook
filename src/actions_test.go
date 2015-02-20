package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_extractTagNames(t *testing.T) {
	expext := []string{"tag1", "tag2", "tag2"}

	values := url.Values{}

	values.Add("tags[]", "tag1")
	values.Add("tags[]", "tag2")
	values.Add("tags[]", "tag2")

	assert.Equal(t, expext, extractTagNames(values))

	values.Del("tags[]")
	values.Set("tags", "tag1,tag2,tag2")

	assert.Equal(t, expext, extractTagNames(values))
}

func Test_buildCreateLogResponse(t *testing.T) {
	now := time.Now()
	tags := []Tag{Tag{1, "tag1"}, Tag{1, "tag2"}}

	logRecord := LogRecord{
		Id:          123,
		Application: Application{1, "TestApp"},
		Level:       5,
		CreatedAt:   &now,
		Message:     "Lorem ipsum",
		Tags:        tags,
	}

	res, _ := buildCreateLogResponse(&logRecord)

	parsed := LogRecordResponse{}
	json.Unmarshal(res, &parsed)

	fmt.Println(logRecord.CreatedAt.Zone())
	fmt.Println(parsed.CreatedAt.Zone())

	assert.Equal(t, 123, parsed.Id)
	assert.Equal(t, "TestApp", parsed.Application)
	assert.Equal(t, 5, parsed.Level)
	assert.EqualValues(t, now, *parsed.CreatedAt)
	assert.Equal(t, "Lorem ipsum", parsed.Message)
	assert.Equal(t, []string{"tag1", "tag2"}, parsed.Tags)
}

func Test_checkCreateLogParams(t *testing.T) {
	assert.Nil(t, checkCreateLogParams("Lorem", "1", []string{"tag1", "tag2"}))
	assert.Nil(t, checkCreateLogParams("Lorem", "1", []string{}))
	assert.Nil(t, checkCreateLogParams("Lorem", "5", []string{}))

	assert.Error(t, checkCreateLogParams("", "1", []string{}))
	assert.Error(t, checkCreateLogParams("Lorem", "", []string{}))
	assert.Error(t, checkCreateLogParams("Lorem", "0", []string{}))
	assert.Error(t, checkCreateLogParams("Lorem", "6", []string{}))
	assert.Error(t, checkCreateLogParams("Lorem", "1", []string{""}))
}

func Test_buildGetLogsResponse(t *testing.T) {
	now := time.Now()
	tags := []Tag{Tag{1, "tag1"}, Tag{1, "tag2"}}

	logRecords := []LogRecord{
		LogRecord{
			Id:          123,
			Application: Application{1, "TestApp"},
			Level:       5,
			CreatedAt:   &now,
			Message:     "Lorem ipsum",
			Tags:        tags,
		},
		LogRecord{
			Id:          456,
			Application: Application{1, "TestApp"},
			Level:       4,
			CreatedAt:   &now,
			Message:     "Dolor sit amet",
			Tags:        tags,
		},
	}

	res, _ := buildGetLogsResponse(&logRecords)

	parsed := LogRecordsResponse{}
	json.Unmarshal(res, &parsed)

	assert.Len(t, parsed, 2)

	for i, response := range parsed {
		assert.Equal(t, logRecords[i].Id, response.Id)
		assert.Equal(t, logRecords[i].Application.Name, response.Application)
		assert.Equal(t, logRecords[i].Level, response.Level)
		assert.EqualValues(t, now, *response.CreatedAt)
		assert.Equal(t, logRecords[i].Message, response.Message)
		assert.Equal(t, []string{"tag1", "tag2"}, response.Tags)
	}
}

func Test_checkGetLogParams(t *testing.T) {
	assert.Nil(t, checkGetLogParams("1", []string{"tag1", "tag2"}, "2014-01-02", "2014-03-04", "1"))
	assert.Nil(t, checkGetLogParams("1", []string{}, "2014-01-02", "2014-03-04", "1"))
	assert.Nil(t, checkGetLogParams("5", []string{}, "2014-01-02", "2014-03-04", "1"))
	assert.Nil(t, checkGetLogParams("1", []string{}, "2014-01-02 03:04:05", "2014-03-04 05:06:07", "1"))

	assert.Error(t, checkGetLogParams("", []string{}, "2014-01-02", "2014-03-04", "1"))
	assert.Error(t, checkGetLogParams("0", []string{}, "2014-01-02", "2014-03-04", "1"))
	assert.Error(t, checkGetLogParams("6", []string{}, "2014-01-02", "2014-03-04", "1"))
	assert.Error(t, checkGetLogParams("1", []string{""}, "2014-01-02", "2014-03-04", "1"))
	assert.Error(t, checkGetLogParams("1", []string{}, "2014-01-022", "2014-03-04", "1"))
	assert.Error(t, checkGetLogParams("1", []string{}, "2014-01-02", "2014-03-044", "1"))
	assert.Error(t, checkGetLogParams("1", []string{}, "2014-01-02", "2014-03-04", "a"))
}
