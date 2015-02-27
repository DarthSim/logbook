package main

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type ModelsTestSuite struct {
	suite.Suite
}

func (suite *ModelsTestSuite) Date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.Local)
}

func (suite *ModelsTestSuite) SetupSuite() {
	config = Config{}

	config.Database.Path = "test.sqlite"
	config.Database.LockTimeout = 1
	config.Database.RetryDelay = 10
	config.Database.MaxOpenConnections = 5
	config.Database.MaxIdleConnections = 5

	config.Log.Path = "test.log"
	config.Log.LogDatabase = false

	config.Auth.User = "test"
	config.Auth.Password = "test"

	initLogger()
}

func (suite *ModelsTestSuite) TearDownSuite() {
	closeLogger()
	os.Remove(absPathToFile(config.Log.Path))
}

func (suite *ModelsTestSuite) SetupTest() {
	initDB()
}

func (suite *ModelsTestSuite) TearDownTest() {
	closeDB()
	os.Remove(absPathToFile(config.Database.Path))
}

func (suite *ModelsTestSuite) Test_createLogRecord() {
	logRecord, err := createLogRecord(
		"apptest",
		"Message one",
		5,
		[]string{"tag1", "tag2"},
	)

	suite.Nil(err)

	suite.Equal("Message one", logRecord.Message)
	suite.Equal(5, logRecord.Level)
	suite.Equal("apptest", logRecord.Application.Name)
	suite.Equal("tag1", logRecord.Tags[0].Name)
	suite.Equal("tag2", logRecord.Tags[1].Name)

	var (
		applications   []Application
		logRecords     []LogRecord
		tags           []Tag
		logRecordsTags []LogRecordTag
	)

	dbmap.Select(&applications, "SELECT * FROM applications")
	dbmap.Select(&logRecords, "SELECT * FROM log_records")
	dbmap.Select(&logRecordsTags, "SELECT * FROM log_records_tags")
	dbmap.Select(&tags, "SELECT * FROM tags")

	suite.Len(applications, 1)
	suite.Equal("apptest", applications[0].Name)

	suite.Len(logRecords, 1)
	suite.Equal(applications[0].Id, logRecords[0].ApplicationId)
	suite.Equal("Message one", logRecords[0].Message)
	suite.Equal(5, logRecords[0].Level)

	suite.Len(tags, 2)
	suite.Equal("tag1", tags[0].Name)
	suite.Equal("tag2", tags[1].Name)

	suite.Len(logRecordsTags, 2)
	suite.Equal(logRecords[0].Id, logRecordsTags[0].LogRecordId)
	suite.Equal(tags[0].Id, logRecordsTags[0].TagId)
	suite.Equal(logRecords[0].Id, logRecordsTags[1].LogRecordId)
	suite.Equal(tags[1].Id, logRecordsTags[1].TagId)
}

func (suite *ModelsTestSuite) Test_createLogRecord_Twice() {
	createLogRecord("apptest", "Message one", 5, []string{"tag1", "tag2"})
	createLogRecord("apptest", "Message two", 3, []string{"tag2", "tag3"})

	var (
		applications   []Application
		logRecords     []LogRecord
		tags           []Tag
		logRecordsTags []LogRecordTag
	)

	dbmap.Select(&applications, "SELECT * FROM applications")
	dbmap.Select(&logRecords, "SELECT * FROM log_records")
	dbmap.Select(&logRecordsTags, "SELECT * FROM log_records_tags")
	dbmap.Select(&tags, "SELECT * FROM tags")

	suite.Len(applications, 1, "Applications should be reusable")
	suite.Equal("apptest", applications[0].Name)

	suite.Len(logRecords, 2)
	suite.Equal(applications[0].Id, logRecords[1].ApplicationId)
	suite.Equal("Message two", logRecords[1].Message)
	suite.Equal(3, logRecords[1].Level)

	suite.Len(tags, 3, "Tags should be reusable")
	suite.Equal("tag2", tags[1].Name)
	suite.Equal("tag3", tags[2].Name)

	suite.Len(logRecordsTags, 4)
	suite.Equal(logRecords[1].Id, logRecordsTags[2].LogRecordId)
	suite.Equal(tags[1].Id, logRecordsTags[2].TagId)
	suite.Equal(logRecords[1].Id, logRecordsTags[3].LogRecordId)
	suite.Equal(tags[2].Id, logRecordsTags[3].TagId)
}

func (suite *ModelsTestSuite) Test_findLogRecords_WithoutTags() {
	applications := []Application{
		Application{0, "testapp1"},
		Application{0, "testapp2"},
	}
	suite.Nil(dbmap.Insert(&applications[0], &applications[1]))

	logRecords := []LogRecord{
		LogRecord{ApplicationId: applications[0].Id, Level: 5, CreatedAt: suite.Date(2014, 01, 01), Message: "Message 1"},
		LogRecord{ApplicationId: applications[1].Id, Level: 5, CreatedAt: suite.Date(2014, 02, 01), Message: "Message 2"},
		LogRecord{ApplicationId: applications[0].Id, Level: 1, CreatedAt: suite.Date(2014, 03, 01), Message: "Message 3"},
		LogRecord{ApplicationId: applications[0].Id, Level: 2, CreatedAt: suite.Date(2014, 04, 01), Message: "Message 4"},
		LogRecord{ApplicationId: applications[0].Id, Level: 5, CreatedAt: suite.Date(2014, 05, 01), Message: "Message 5"},
		LogRecord{ApplicationId: applications[0].Id, Level: 5, CreatedAt: suite.Date(2014, 06, 01), Message: "Message 6"},
	}
	suite.Nil(dbmap.Insert(&logRecords[0], &logRecords[1], &logRecords[2],
		&logRecords[3], &logRecords[4], &logRecords[5]))

	tags := []Tag{Tag{0, "tag1"}, Tag{0, "tag2"}}
	suite.Nil(dbmap.Insert(&tags[0], &tags[1]))

	logRecordTags := []LogRecordTag{
		LogRecordTag{logRecords[3].Id, tags[0].Id},
		LogRecordTag{logRecords[3].Id, tags[1].Id},
	}
	suite.Nil(dbmap.Insert(&logRecordTags[0], &logRecordTags[1]))

	foundLogRecords, err := findLogRecords("testapp1", 2, []string{},
		suite.Date(2014, 01, 15), suite.Date(2014, 05, 15), 1)

	suite.Nil(err)
	suite.Len(foundLogRecords, 2)

	suite.Equal(logRecords[4].Id, foundLogRecords[0].Id)
	suite.Equal(applications[0], foundLogRecords[0].Application)
	suite.Equal(logRecords[4].Level, foundLogRecords[0].Level)
	suite.Equal(logRecords[4].CreatedAt, foundLogRecords[0].CreatedAt)
	suite.Equal(logRecords[4].Message, foundLogRecords[0].Message)
	suite.Empty(foundLogRecords[0].Tags)

	suite.Equal(logRecords[3].Id, foundLogRecords[1].Id)
	suite.Equal(applications[0], foundLogRecords[1].Application)
	suite.Equal(logRecords[3].Level, foundLogRecords[1].Level)
	suite.Equal(logRecords[3].CreatedAt, foundLogRecords[1].CreatedAt)
	suite.Equal(logRecords[3].Message, foundLogRecords[1].Message)
	suite.Equal(tags, foundLogRecords[1].Tags)
}

func (suite *ModelsTestSuite) Test_findLogRecords_WithTags() {
	application := Application{0, "testapp1"}
	suite.Nil(dbmap.Insert(&application))

	tags := []Tag{Tag{0, "tag1"}, Tag{0, "tag2"}, Tag{0, "tag3"}}
	suite.Nil(dbmap.Insert(&tags[0], &tags[1], &tags[2]))

	logRecords := []LogRecord{
		LogRecord{ApplicationId: application.Id, Level: 5, CreatedAt: suite.Date(2014, 02, 01), Message: "Message 1"},
		LogRecord{ApplicationId: application.Id, Level: 5, CreatedAt: suite.Date(2014, 02, 01), Message: "Message 2"},
		LogRecord{ApplicationId: application.Id, Level: 5, CreatedAt: suite.Date(2014, 02, 01), Message: "Message 3"},
	}
	suite.Nil(dbmap.Insert(&logRecords[0], &logRecords[1], &logRecords[2]))

	logRecordTags := []LogRecordTag{
		LogRecordTag{logRecords[0].Id, tags[0].Id},
		LogRecordTag{logRecords[0].Id, tags[1].Id},
		LogRecordTag{logRecords[0].Id, tags[2].Id},
		LogRecordTag{logRecords[1].Id, tags[0].Id},
		LogRecordTag{logRecords[1].Id, tags[1].Id},
		LogRecordTag{logRecords[2].Id, tags[1].Id},
		LogRecordTag{logRecords[2].Id, tags[2].Id},
	}
	suite.Nil(dbmap.Insert(&logRecordTags[0], &logRecordTags[1], &logRecordTags[2],
		&logRecordTags[3], &logRecordTags[4], &logRecordTags[5], &logRecordTags[6]))

	foundLogRecords, err := findLogRecords("testapp1", 2, []string{"tag1", "tag2"},
		suite.Date(2014, 01, 15), suite.Date(2014, 05, 15), 1)

	suite.Nil(err)
	suite.Len(foundLogRecords, 2)

	suite.Equal(logRecords[1].Id, foundLogRecords[0].Id)
	suite.Equal(logRecords[0].Id, foundLogRecords[1].Id)
}

func (suite *ModelsTestSuite) Test_findLogRecords_Pagination() {
	application := Application{0, "testapp1"}
	suite.Nil(dbmap.Insert(&application))

	logRecords := make([]LogRecord, 110)
	for i := 0; i < 110; i++ {
		logRecords[i] = LogRecord{ApplicationId: application.Id, Level: 5, CreatedAt: time.Now(), Message: "Message 1"}
		dbmap.Insert(&logRecords[i])
	}

	foundLogRecords, err := findLogRecords("testapp1", 2, []string{},
		suite.Date(2014, 01, 01), suite.Date(3014, 01, 01), 1)

	suite.Nil(err)
	suite.Len(foundLogRecords, 100)
	suite.Equal(logRecords[109].Id, foundLogRecords[0].Id)
	suite.Equal(logRecords[10].Id, foundLogRecords[99].Id)

	foundLogRecords, err = findLogRecords("testapp1", 2, []string{},
		suite.Date(2014, 01, 01), suite.Date(3014, 01, 01), 2)

	suite.Nil(err)
	suite.Len(foundLogRecords, 10)
	suite.Equal(logRecords[9].Id, foundLogRecords[0].Id)
	suite.Equal(logRecords[0].Id, foundLogRecords[9].Id)
}

func TestModels(t *testing.T) {
	suite.Run(t, new(ModelsTestSuite))
}
