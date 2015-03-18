package main

import (
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Models", func() {
	BeforeEach(func() {
		initDB()
	})

	AfterEach(func() {
		closeDB()
		os.Remove(absPathToFile(config.Database.Path))
	})

	Describe("createLogRecord", func() {
		var logRecord LogRecord

		JustBeforeEach(func() {
			var err error

			logRecord, err = createLogRecord(
				"apptest",
				"Message one",
				5,
				[]string{"tag1", "tag2"},
			)

			Expect(err).NotTo(HaveOccurred())
		})

		It("should create log record with application and tags", func() {
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

			Expect(applications).To(HaveLen(1))
			Expect(applications[0].Name).To(Equal("apptest"))

			Expect(logRecords).To(HaveLen(1))
			Expect(logRecords[0].ApplicationId).To(Equal(applications[0].Id))
			Expect(logRecords[0].Message).To(Equal("Message one"))
			Expect(logRecords[0].Level).To(Equal(5))

			Expect(tags).To(HaveLen(2))
			Expect(tags[0].Name).To(Equal("tag1"))
			Expect(tags[1].Name).To(Equal("tag2"))

			Expect(logRecordsTags).To(HaveLen(2))
			Expect(logRecordsTags[0].LogRecordId).To(Equal(logRecords[0].Id))
			Expect(logRecordsTags[0].TagId).To(Equal(tags[0].Id))
			Expect(logRecordsTags[1].LogRecordId).To(Equal(logRecords[0].Id))
			Expect(logRecordsTags[1].TagId).To(Equal(tags[1].Id))
		})

		It("should return created log record", func() {
			Expect(logRecord.Message).To(Equal("Message one"))
			Expect(logRecord.Level).To(Equal(5))
			Expect(logRecord.Application.Name).To(Equal("apptest"))
			Expect(logRecord.Tags[0].Name).To(Equal("tag1"))
			Expect(logRecord.Tags[1].Name).To(Equal("tag2"))
		})

		Context("when used many times", func() {
			It("should reuse application and tags", func() {
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

				Expect(applications).To(HaveLen(1))
				Expect(applications[0].Name).To(Equal("apptest"))

				Expect(logRecords).To(HaveLen(2))
				Expect(logRecords[1].ApplicationId).To(Equal(applications[0].Id))
				Expect(logRecords[1].Message).To(Equal("Message two"))
				Expect(logRecords[1].Level).To(Equal(3))

				Expect(tags).To(HaveLen(3))
				Expect(tags[1].Name).To(Equal("tag2"))
				Expect(tags[2].Name).To(Equal("tag3"))

				Expect(logRecordsTags).To(HaveLen(4))
				Expect(logRecordsTags[2].LogRecordId).To(Equal(logRecords[1].Id))
				Expect(logRecordsTags[2].TagId).To(Equal(tags[1].Id))
				Expect(logRecordsTags[3].LogRecordId).To(Equal(logRecords[1].Id))
				Expect(logRecordsTags[3].TagId).To(Equal(tags[2].Id))
			})
		})
	})

	Describe("findLogRecords", func() {
		buildDate := func(year int, month time.Month, day int) time.Time {
			return time.Date(year, month, day, 0, 0, 0, 0, time.Local)
		}

		It("should return return log records filtered by app level and time", func() {
			applications := []Application{
				Application{0, "testapp1"},
				Application{0, "testapp2"},
			}
			Expect(
				dbmap.Insert(&applications[0], &applications[1]),
			).To(Succeed())

			logRecords := []LogRecord{
				LogRecord{ApplicationId: applications[0].Id, Level: 5, CreatedAt: buildDate(2014, 01, 01), Message: "Message 1"},
				LogRecord{ApplicationId: applications[1].Id, Level: 5, CreatedAt: buildDate(2014, 02, 01), Message: "Message 2"},
				LogRecord{ApplicationId: applications[0].Id, Level: 1, CreatedAt: buildDate(2014, 03, 01), Message: "Message 3"},
				LogRecord{ApplicationId: applications[0].Id, Level: 2, CreatedAt: buildDate(2014, 04, 01), Message: "Message 4"},
				LogRecord{ApplicationId: applications[0].Id, Level: 5, CreatedAt: buildDate(2014, 05, 01), Message: "Message 5"},
				LogRecord{ApplicationId: applications[0].Id, Level: 5, CreatedAt: buildDate(2014, 06, 01), Message: "Message 6"},
			}
			Expect(
				dbmap.Insert(&logRecords[0], &logRecords[1], &logRecords[2],
					&logRecords[3], &logRecords[4], &logRecords[5]),
			).To(Succeed())

			tags := []Tag{Tag{0, "tag1"}, Tag{0, "tag2"}}
			Expect(dbmap.Insert(&tags[0], &tags[1])).To(Succeed())

			logRecordTags := []LogRecordTag{
				LogRecordTag{logRecords[3].Id, tags[0].Id},
				LogRecordTag{logRecords[3].Id, tags[1].Id},
			}
			Expect(
				dbmap.Insert(&logRecordTags[0], &logRecordTags[1]),
			).To(Succeed())

			foundLogRecords, err := findLogRecords("testapp1", 2, []string{},
				buildDate(2014, 01, 15), buildDate(2014, 05, 15), 1)

			Expect(err).NotTo(HaveOccurred())
			Expect(foundLogRecords).To(HaveLen(2))

			Expect(foundLogRecords[0].Id).To(Equal(logRecords[4].Id))
			Expect(foundLogRecords[0].Application).To(Equal(applications[0]))
			Expect(foundLogRecords[0].Level).To(Equal(logRecords[4].Level))
			Expect(foundLogRecords[0].CreatedAt.UTC()).To(Equal(logRecords[4].CreatedAt.UTC()))
			Expect(foundLogRecords[0].Message).To(Equal(logRecords[4].Message))
			Expect(foundLogRecords[0].Tags).To(BeEmpty())

			Expect(foundLogRecords[1].Id).To(Equal(logRecords[3].Id))
			Expect(foundLogRecords[1].Application).To(Equal(applications[0]))
			Expect(foundLogRecords[1].Level).To(Equal(logRecords[3].Level))
			Expect(foundLogRecords[1].CreatedAt.UTC()).To(Equal(logRecords[3].CreatedAt.UTC()))
			Expect(foundLogRecords[1].Message).To(Equal(logRecords[3].Message))
			Expect(foundLogRecords[1].Tags).To(Equal(tags))
		})

		Context("with tags", func() {
			It("should also filter log records by tags", func() {
				application := Application{0, "testapp1"}
				Expect(dbmap.Insert(&application)).To(Succeed())

				tags := []Tag{Tag{0, "tag1"}, Tag{0, "tag2"}, Tag{0, "tag3"}}
				Expect(dbmap.Insert(&tags[0], &tags[1], &tags[2])).To(Succeed())

				logRecords := []LogRecord{
					LogRecord{ApplicationId: application.Id, Level: 5, CreatedAt: buildDate(2014, 02, 01), Message: "Message 1"},
					LogRecord{ApplicationId: application.Id, Level: 5, CreatedAt: buildDate(2014, 02, 01), Message: "Message 2"},
					LogRecord{ApplicationId: application.Id, Level: 5, CreatedAt: buildDate(2014, 02, 01), Message: "Message 3"},
				}
				Expect(
					dbmap.Insert(&logRecords[0], &logRecords[1], &logRecords[2]),
				).To(Succeed())

				logRecordTags := []LogRecordTag{
					LogRecordTag{logRecords[0].Id, tags[0].Id},
					LogRecordTag{logRecords[0].Id, tags[1].Id},
					LogRecordTag{logRecords[0].Id, tags[2].Id},
					LogRecordTag{logRecords[1].Id, tags[0].Id},
					LogRecordTag{logRecords[1].Id, tags[1].Id},
					LogRecordTag{logRecords[2].Id, tags[1].Id},
					LogRecordTag{logRecords[2].Id, tags[2].Id},
				}
				Expect(
					dbmap.Insert(&logRecordTags[0], &logRecordTags[1], &logRecordTags[2],
						&logRecordTags[3], &logRecordTags[4], &logRecordTags[5],
						&logRecordTags[6]),
				).To(Succeed())

				foundLogRecords, err := findLogRecords("testapp1", 2, []string{"tag1", "tag2"},
					buildDate(2014, 01, 15), buildDate(2014, 05, 15), 1)

				Expect(err).NotTo(HaveOccurred())

				Expect(foundLogRecords).To(HaveLen(2))
				Expect(foundLogRecords[0].Id).To(Equal(logRecords[1].Id))
				Expect(foundLogRecords[1].Id).To(Equal(logRecords[0].Id))
			})
		})

		Context("with pagination", func() {
			It("should paginate results", func() {
				application := Application{0, "testapp1"}
				Expect(dbmap.Insert(&application)).To(Succeed())

				logRecords := make([]LogRecord, 110)
				for i := 0; i < 110; i++ {
					logRecords[i] = LogRecord{ApplicationId: application.Id, Level: 5,
						CreatedAt: time.Now(), Message: "Message 1"}
					Expect(dbmap.Insert(&logRecords[i])).To(Succeed())
				}

				foundLogRecords, err := findLogRecords("testapp1", 2, []string{},
					buildDate(2014, 01, 01), buildDate(3014, 01, 01), 1)

				Expect(err).NotTo(HaveOccurred())

				Expect(foundLogRecords).To(HaveLen(100))
				Expect(foundLogRecords[0].Id).To(Equal(logRecords[109].Id))
				Expect(foundLogRecords[99].Id).To(Equal(logRecords[10].Id))

				foundLogRecords, err = findLogRecords("testapp1", 2, []string{},
					buildDate(2014, 01, 01), buildDate(3014, 01, 01), 2)

				Expect(err).NotTo(HaveOccurred())

				Expect(foundLogRecords).To(HaveLen(10))
				Expect(foundLogRecords[0].Id).To(Equal(logRecords[9].Id))
				Expect(foundLogRecords[9].Id).To(Equal(logRecords[0].Id))
			})
		})
	})
})
