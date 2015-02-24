package main

import (
	"testing"
	// "time"

	// "database/sql"
	// "reflect"
	// "strings"
	// "time"
	// "fmt"

	// "github.com/coopernurse/gorp"
	// "github.com/mattn/go-sqlite3"
)

func Test_createLogRecord_good(t *testing.T) {
	prepareDb()
	record, err := createLogRecord("demo", "msg", 5, []string{"foo", "bar"})

	if err != nil {
		t.Error()
	}

	if record.Message != "msg" {
		t.Error()
	}
	if record.Level != 5 {
		t.Error()
	}
	if len(record.Tags) != 2 {
		t.Error()
	}
	if record.Tags[0].Name != "foo" {
		t.Error()
	}
	if record.Tags[1].Name != "bar" {
		t.Error()
	}
}

func prepareDb() {
	config = Config{}
	config.Database.Path = "test.sqlite"
	config.Database.LockTimeout = 1
	config.Database.RetryDelay = 10
	config.Database.MaxIdleConnections = 5

	config.Log.Path = "test.log"
	config.Log.LogDatabase = false

	initDB()
}
