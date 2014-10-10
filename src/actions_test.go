package main

import (
	"encoding/json"
	"testing"
	"time"
)

func Test_buildCreateLogResponse(t *testing.T) {
	now := time.Now()
	tags := []Tag{Tag{}}

	obj := LogRecord{
		Id:          123,
		Application: Application{},
		Level:       5,
		CreatedAt:   &now,
		Message:     "fff",
		Tags:        tags,
	}

	res, _ := buildCreateLogResponse(&obj)

	parsed := LogRecord{}
	json.Unmarshal(res, &parsed)

	if parsed.Id != 123 {
		t.Error()
	}

	if parsed.Message != "fff" {
		t.Error()
	}

	if parsed.Level != 5 {
		t.Error()
	}
}
