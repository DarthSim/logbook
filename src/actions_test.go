package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	router   *gin.Engine
	response *httptest.ResponseRecorder
)

func sendRequest(method, path string, body ...string) (err error) {
	var req *http.Request

	response = httptest.NewRecorder()

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

	router.ServeHTTP(response, req)

	return nil
}

var _ = Describe("Actions", func() {
	var query string

	BeforeEach(func() {
		gin.SetMode(gin.TestMode)
		router = setupRouter()
	})

	AssertSuccess := func() {
		It("should respond with 200", func() {
			Expect(response.Code).To(Equal(200))
		})
	}

	AssertUnprocessable := func() {
		It("should respond with 422", func() {
			Expect(response.Code).To(Equal(422))
		})

		It("should respond with error", func() {
			parsedRes := ErrorResponse{}
			Expect(
				json.Unmarshal(response.Body.Bytes(), &parsedRes),
			).To(Succeed())

			Expect(parsedRes.Error).NotTo(BeEmpty())
		})
	}

	Describe("/:application/put", func() {
		BeforeEach(func() {
			query = "message=Lorem%20ipsum&level=1&tags=tag1,tag2"
		})

		JustBeforeEach(func() {
			Expect(
				sendRequest("POST", "/apptest/put", query),
			).To(Succeed())
		})

		AssertSuccess()

		It("should respond with created log record", func() {
			parsedRes := LogRecord{}
			Expect(
				json.Unmarshal(response.Body.Bytes(), &parsedRes),
			).To(Succeed())

			Expect(parsedRes.Message).To(Equal("Lorem ipsum"))
			Expect(parsedRes.Level).To(Equal(1))
			Expect(parsedRes.CreatedAt).To(BeTemporally("~", time.Now(), time.Second))
			Expect(parsedRes.Tags).To(ConsistOf([]string{"tag1", "tag2"}))
		})

		Context("with duplicate tags", func() {
			BeforeEach(func() {
				query = "message=Lorem%20ipsum&level=1&tags=tag1,tag2,tag1"
			})

			AssertSuccess()

			It("should remove duplicate tags", func() {
				parsedRes := LogRecord{}
				Expect(
					json.Unmarshal(response.Body.Bytes(), &parsedRes),
				).To(Succeed())
				Expect(parsedRes.Tags).To(ConsistOf([]string{"tag1", "tag2"}))
			})
		})

		Context("without tags", func() {
			BeforeEach(func() {
				query = "message=Lorem%20ipsum&level=1"
			})
			AssertSuccess()
		})

		Context("with level 5", func() {
			BeforeEach(func() {
				query = "message=Lorem%20ipsum&level=5&tags=tag1,tag2"
			})
			AssertSuccess()
		})

		Context("without message", func() {
			BeforeEach(func() {
				query = "message=level=5&tags=tag1,tag2"
			})
			AssertUnprocessable()
		})

		Context("without level", func() {
			BeforeEach(func() {
				query = "message=Lorem%20ipsum&level=&tags=tag1,tag2"
			})
			AssertUnprocessable()
		})

		Context("with level < 0", func() {
			BeforeEach(func() {
				query = "message=Lorem%20ipsum&level=-1&tags=tag1,tag2"
			})
			AssertUnprocessable()
		})

		Context("with level > 6", func() {
			BeforeEach(func() {
				query = "message=Lorem%20ipsum&level=6&tags=tag1,tag2"
			})
			AssertUnprocessable()
		})

		Context("with empty string tag", func() {
			BeforeEach(func() {
				query = "message=Lorem%20ipsum&level=1&tags=tag1,,tag2"
			})
			AssertUnprocessable()
		})

		Context("with created_at", func() {
			BeforeEach(func() {
				query = "message=Lorem%20ipsum&level=1&tags=tag1,tag2&created_at=2015-10-16T08:10:11.123"
			})

			It("should create log record with provided datetime", func() {
				parsedRes := LogRecord{}
				Expect(
					json.Unmarshal(response.Body.Bytes(), &parsedRes),
				).To(Succeed())

				Expect(parsedRes.CreatedAt.Local().Truncate(time.Millisecond)).To(
					Equal(time.Date(2015, 10, 16, 8, 10, 11, 123000000, time.Local)),
				)
			})
		})

		Context("with invalid created_at", func() {
			BeforeEach(func() {
				query = "message=Lorem%20ipsum&level=1&tags=tag1,tag2&created_at=2015-10-16T08:10:111"
			})
			AssertUnprocessable()
		})
	})

	Describe("/:application/get", func() {
		generateLogRecord := func(application, message string, level int, tags ...string) {
			logRecord := LogRecord{
				Message: message,
				Level:   level,
				Tags:    tags,
			}
			Expect(
				storage.SaveLogRecord(application, &logRecord),
			).To(Succeed())
		}

		BeforeEach(func() {
			generateLogRecord("testapp1", "Message one", 1, "tag2, tag3", "tag4")
			generateLogRecord("testapp1", "Message two", 2, "tag2", "tag3", "tag4")
			generateLogRecord("testapp1", "Message three", 3, "tag3", "tag4", "tag5")
			generateLogRecord("testapp1", "Message four", 3, "tag1", "tag2", "tag3")
			generateLogRecord("testapp2", "Message five", 3, "tag3", "tag4")

			query = fmt.Sprintf(
				"level=2&tags=tag3,tag4&start_time=%v&end_time=%v&page=1",
				time.Now().Format("2006-01-02"),
				time.Now().Format("2006-01-02"),
			)
		})

		JustBeforeEach(func() {
			Expect(
				sendRequest("GET", "/testapp1/get?"+query),
			).To(Succeed())
		})

		AssertSuccess()

		It("should respond with found log records", func() {
			parsedRes := LogRecords{}
			Expect(
				json.Unmarshal(response.Body.Bytes(), &parsedRes),
			).To(Succeed())

			Expect(parsedRes).To(HaveLen(2))

			Expect(parsedRes[0].Message).To(Equal("Message two"))
			Expect(parsedRes[0].Level).To(Equal(2))
			Expect(parsedRes[0].CreatedAt).To(BeAssignableToTypeOf(time.Now()))
			Expect(parsedRes[0].Tags).To(ConsistOf("tag2", "tag3", "tag4"))

			Expect(parsedRes[1].Message).To(Equal("Message three"))
			Expect(parsedRes[1].Level).To(Equal(3))
			Expect(parsedRes[1].CreatedAt).To(BeAssignableToTypeOf(time.Now()))
			Expect(parsedRes[1].Tags).To(ConsistOf("tag3", "tag4", "tag5"))
		})

		Context("without tags", func() {
			BeforeEach(func() {
				query = "level=1&start_time=2006-01-02&end_time=2006-01-02&page=1"
			})
			AssertSuccess()
		})

		Context("with level 5", func() {
			BeforeEach(func() {
				query = "level=5&tags=tag3,tag4&start_time=2006-01-02&end_time=2006-01-02&page=1"
			})
			AssertSuccess()
		})

		Context("with datetime as start_time and end_time", func() {
			BeforeEach(func() {
				query = "level=1&tags=tag3,tag4&start_time=2006-01-02T01:02:03&end_time=2006-01-02T01:02:03&page=1"
			})
			AssertSuccess()
		})

		Context("without page", func() {
			BeforeEach(func() {
				query = "level=1&tags=tag3,tag4&start_time=2006-01-02T01:02:03&end_time=2006-01-02T01:02:03"
			})
			AssertSuccess()
		})

		Context("without level", func() {
			BeforeEach(func() {
				query = "tags=tag3,tag4&start_time=2006-01-02&end_time=2006-01-02&page=1"
			})
			AssertUnprocessable()
		})

		Context("with level < 0", func() {
			BeforeEach(func() {
				query = "level=-1tags=tag3,tag4&start_time=2006-01-02&end_time=2006-01-02&page=1"
			})
			AssertUnprocessable()
		})

		Context("with level > 5", func() {
			BeforeEach(func() {
				query = "level=6tags=tag3,tag4&start_time=2006-01-02&end_time=2006-01-02&page=1"
			})
			AssertUnprocessable()
		})

		Context("with empty string tag", func() {
			BeforeEach(func() {
				query = "level=1tags=tag3,,tag4&start_time=2006-01-02&end_time=2006-01-02&page=1"
			})
			AssertUnprocessable()
		})

		Context("with invalid start_time", func() {
			BeforeEach(func() {
				query = "level=1tags=tag3,tag4&start_time=2006-01-022&end_time=2006-01-02&page=1"
			})
			AssertUnprocessable()
		})

		Context("with invalid end_time", func() {
			BeforeEach(func() {
				query = "level=1tags=tag3,tag4&start_time=2006-01-02&end_time=2006-01-022&page=1"
			})
			AssertUnprocessable()
		})

		Context("with invalid page", func() {
			BeforeEach(func() {
				query = "level=1tags=tag3,tag4&start_time=2006-01-02&end_time=2006-01-02&page=a"
			})
			AssertUnprocessable()
		})
	})
})
