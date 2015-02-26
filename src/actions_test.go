package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

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

func Test_checkGetLogParams(t *testing.T) {
	assert.Nil(t, checkGetLogParams("1", []string{"tag1", "tag2"}, "2014-01-02",
		"2014-03-04", "1"))
	assert.Nil(t, checkGetLogParams("1", []string{}, "2014-01-02",
		"2014-03-04", "1"))
	assert.Nil(t, checkGetLogParams("5", []string{}, "2014-01-02",
		"2014-03-04", "1"))
	assert.Nil(t, checkGetLogParams("1", []string{}, "2014-01-02 03:04:05",
		"2014-03-04 05:06:07", "1"))

	assert.Error(t, checkGetLogParams("", []string{}, "2014-01-02",
		"2014-03-04", "1"))
	assert.Error(t, checkGetLogParams("0", []string{}, "2014-01-02",
		"2014-03-04", "1"))
	assert.Error(t, checkGetLogParams("6", []string{}, "2014-01-02",
		"2014-03-04", "1"))
	assert.Error(t, checkGetLogParams("1", []string{""}, "2014-01-02",
		"2014-03-04", "1"))
	assert.Error(t, checkGetLogParams("1", []string{}, "2014-01-022",
		"2014-03-04", "1"))
	assert.Error(t, checkGetLogParams("1", []string{}, "2014-01-02",
		"2014-03-044", "1"))
	assert.Error(t, checkGetLogParams("1", []string{}, "2014-01-02",
		"2014-03-04", "a"))
}

// Note:
// There is a way to use httptest.Server to start test server and send real
// reuests. But I prefer not to do this and use response recorder with router.
// If anybody thinks I am wrong - plase let me know.

type ActionsTestSuite struct {
	suite.Suite
	Router   *mux.Router
	Response *httptest.ResponseRecorder
}

func (suite *ActionsTestSuite) SetupSuite() {
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

	suite.Router = setupRouter()
}

func (suite *ActionsTestSuite) TearDownSuite() {
	closeLogger()
	os.Remove(absPathToFile(config.Log.Path))
}

func (suite *ActionsTestSuite) SetupTest() {
	initDB()
}

func (suite *ActionsTestSuite) TearDownTest() {
	closeDB()
	os.Remove(absPathToFile(config.Database.Path))
}

func (suite *ActionsTestSuite) SendRequest(method, path string, body ...string) (err error) {
	var req *http.Request

	suite.Response = httptest.NewRecorder()

	if method == "POST" {
		reqBody := strings.NewReader(body[0])
		req, err = http.NewRequest(method, "http://logbook.test"+path, reqBody)
	} else {
		req, err = http.NewRequest(method, "http://logbook.test"+path, nil)
	}

	if err != nil {
		return err
	}

	req.SetBasicAuth(config.Auth.User, config.Auth.Password)

	if method == "POST" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	suite.Router.ServeHTTP(suite.Response, req)
	return nil
}

func (suite *ActionsTestSuite) Test_createLogHandler_Success() {
	suite.Nil(suite.SendRequest(
		"POST",
		"/apptest/put",
		"message=Lorem%20ipsum&level=5&tags=tag1,tag2",
	))

	suite.Equal(200, suite.Response.Code)

	parsedRes := LogRecordResponse{}
	suite.Nil(
		json.Unmarshal(suite.Response.Body.Bytes(), &parsedRes),
	)

	suite.IsType(int64(1), parsedRes.Id)
	suite.Equal("apptest", parsedRes.Application)
	suite.Equal("Lorem ipsum", parsedRes.Message)
	suite.Equal(5, parsedRes.Level)
	suite.IsType(time.Now(), *parsedRes.CreatedAt)
	suite.Equal([]string{"tag1", "tag2"}, parsedRes.Tags)
}

func (suite *ActionsTestSuite) Test_createLogHandler_Failure() {
	suite.Nil(suite.SendRequest(
		"POST",
		"/apptest/put",
		"message=&level=5&tags=tag1,tag2",
	))

	suite.Equal(422, suite.Response.Code)

	parsedRes := ErrorResponse{}
	suite.Nil(
		json.Unmarshal(suite.Response.Body.Bytes(), &parsedRes),
	)

	suite.NotEmpty(parsedRes.Error)
}

func (suite *ActionsTestSuite) Test_getLogsHandler_Success() {
	createLogRecord("testapp1", "Message one", 1, []string{"tag2, tag3", "tag4"})
	createLogRecord("testapp1", "Message two", 2, []string{"tag2", "tag3", "tag4"})
	createLogRecord("testapp1", "Message three", 3, []string{"tag3", "tag4", "tag5"})
	createLogRecord("testapp1", "Message four", 3, []string{"tag1", "tag2", "tag3"})
	createLogRecord("testapp2", "Message five", 3, []string{"tag3", "tag4"})

	suite.Nil(suite.SendRequest(
		"GET",
		fmt.Sprintf(
			"/testapp1/get?level=2&tags=tag3,tag4&start_time=%v&end_time=%v",
			time.Now().Format("2006-01-02"),
			time.Now().Format("2006-01-02"),
		),
	))

	suite.Equal(200, suite.Response.Code)

	parsedRes := LogRecordsResponse{}
	suite.Nil(
		json.Unmarshal(suite.Response.Body.Bytes(), &parsedRes),
	)

	suite.Len(parsedRes, 2)

	if len(parsedRes) == 2 {
		suite.IsType(int64(1), parsedRes[0].Id)
		suite.Equal("testapp1", parsedRes[0].Application)
		suite.Equal("Message two", parsedRes[0].Message)
		suite.Equal(2, parsedRes[0].Level)
		suite.IsType(time.Now(), *parsedRes[0].CreatedAt)
		suite.Contains(parsedRes[0].Tags, "tag2")
		suite.Contains(parsedRes[0].Tags, "tag3")
		suite.Contains(parsedRes[0].Tags, "tag4")

		suite.IsType(int64(1), parsedRes[1].Id)
		suite.Equal("testapp1", parsedRes[1].Application)
		suite.Equal("Message three", parsedRes[1].Message)
		suite.Equal(3, parsedRes[1].Level)
		suite.IsType(time.Now(), *parsedRes[1].CreatedAt)
		suite.Contains(parsedRes[1].Tags, "tag3")
		suite.Contains(parsedRes[1].Tags, "tag4")
		suite.Contains(parsedRes[1].Tags, "tag5")
	}
}

func (suite *ActionsTestSuite) Test_getLogsHandler_Failure() {
	createLogRecord("testapp1", "Message one", 1, []string{"tag2, tag3", "tag4"})

	suite.Nil(suite.SendRequest(
		"GET",
		fmt.Sprintf(
			"/testapp1/get?level=6&start_time=%v&end_time=%v",
			time.Now().Format("2006-01-02"),
			time.Now().Format("2006-01-02"),
		),
	))

	suite.Equal(422, suite.Response.Code)

	parsedRes := ErrorResponse{}
	suite.Nil(
		json.Unmarshal(suite.Response.Body.Bytes(), &parsedRes),
	)

	suite.NotEmpty(parsedRes.Error)
}

func TestActions(t *testing.T) {
	suite.Run(t, new(ActionsTestSuite))
}
