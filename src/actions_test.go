package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	router   *mux.Router
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
		router = setupRouter()

		initDB()
	})

	AfterEach(func() {
		closeDB()
		os.Remove(absPathToFile(config.Database.Path))
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
			parsedRes := LogRecordResponse{}
			Expect(
				json.Unmarshal(response.Body.Bytes(), &parsedRes),
			).To(Succeed())

			Expect(parsedRes.Id).To(BeAssignableToTypeOf(int64(1)))
			Expect(parsedRes.Application).To(Equal("apptest"))
			Expect(parsedRes.Message).To(Equal("Lorem ipsum"))
			Expect(parsedRes.Level).To(Equal(1))
			Expect(parsedRes.CreatedAt).To(BeAssignableToTypeOf(time.Now()))
			Expect(parsedRes.Tags).To(Equal([]string{"tag1", "tag2"}))
		})

		Context("without tags", func() {
			BeforeEach(func() {
				query = "message=Lorem%20ipsum&level=1"
			})
			AssertSuccess()
		})

		Context("with tags as array", func() {
			BeforeEach(func() {
				query = "message=Lorem%20ipsum&level=1&tags[]=tag1&tags[]=tag2"
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

		Context("with level < 1", func() {
			BeforeEach(func() {
				query = "message=Lorem%20ipsum&level=0&tags=tag1,tag2"
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
	})

	Describe("/:application/get", func() {
		BeforeEach(func() {
			createLogRecord("testapp1", "Message one", 1, []string{"tag2, tag3", "tag4"})
			createLogRecord("testapp1", "Message two", 2, []string{"tag2", "tag3", "tag4"})
			createLogRecord("testapp1", "Message three", 3, []string{"tag3", "tag4", "tag5"})
			createLogRecord("testapp1", "Message four", 3, []string{"tag1", "tag2", "tag3"})
			createLogRecord("testapp2", "Message five", 3, []string{"tag3", "tag4"})

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
			parsedRes := LogRecordsResponse{}
			Expect(
				json.Unmarshal(response.Body.Bytes(), &parsedRes),
			).To(Succeed())

			Expect(parsedRes).To(HaveLen(2))

			Expect(parsedRes[0].Id).To(BeAssignableToTypeOf(int64(1)))
			Expect(parsedRes[0].Application).To(Equal("testapp1"))
			Expect(parsedRes[0].Message).To(Equal("Message three"))
			Expect(parsedRes[0].Level).To(Equal(3))
			Expect(parsedRes[0].CreatedAt).To(BeAssignableToTypeOf(time.Now()))
			Expect(parsedRes[0].Tags).To(ConsistOf("tag3", "tag4", "tag5"))

			Expect(parsedRes[1].Id).To(BeAssignableToTypeOf(int64(1)))
			Expect(parsedRes[1].Application).To(Equal("testapp1"))
			Expect(parsedRes[1].Message).To(Equal("Message two"))
			Expect(parsedRes[1].Level).To(Equal(2))
			Expect(parsedRes[1].CreatedAt).To(BeAssignableToTypeOf(time.Now()))
			Expect(parsedRes[1].Tags).To(ConsistOf("tag2", "tag3", "tag4"))
		})

		Context("without tags", func() {
			BeforeEach(func() {
				query = "level=1&start_time=2006-01-02&end_time=2006-01-02&page=1"
			})
			AssertSuccess()
		})

		Context("with tags as array", func() {
			BeforeEach(func() {
				query = "level=1&start_time=2006-01-02&end_time=2006-01-02&tags[]=tag1&tags[]=tag2&page=1"
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
				query = "level=1&tags=tag3,tag4&start_time=2006-01-02%2001:02:03&end_time=2006-01-02%2001:02:03&page=1"
			})
			AssertSuccess()
		})

		Context("without page", func() {
			BeforeEach(func() {
				query = "level=1&tags=tag3,tag4&start_time=2006-01-02%2001:02:03&end_time=2006-01-02%2001:02:03"
			})
			AssertSuccess()
		})

		Context("without level", func() {
			BeforeEach(func() {
				query = "tags=tag3,tag4&start_time=2006-01-02&end_time=2006-01-02&page=1"
			})
			AssertUnprocessable()
		})

		Context("with level < 1", func() {
			BeforeEach(func() {
				query = "level=0tags=tag3,tag4&start_time=2006-01-02&end_time=2006-01-02&page=1"
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
