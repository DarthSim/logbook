package main

import (
	"fmt"
	"time"

	"github.com/tecbot/gorocksdb"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Storage", func() {
	decodeLogRecord := func(data []byte) (logRecord LogRecord) {
		err := logRecord.Decode(data)
		Expect(err).NotTo(HaveOccurred())
		return
	}

	lastKey := func(cf *gorocksdb.ColumnFamilyHandle) []byte {
		it := storage.db.NewIteratorCF(gorocksdb.NewDefaultReadOptions(), cf)
		it.SeekToLast()

		if string(it.Key().Data()) == "::seq::" {
			it.Prev()
		}

		return it.Key().Data()
	}

	Describe("#SaveLogRecord", func() {
		var logRecord LogRecord

		BeforeEach(func() {
			logRecord = LogRecord{
				Message:   "Message one",
				Level:     3,
				Tags:      []string{"tag1", "tag2"},
				CreatedAt: time.Date(2015, 1, 2, 3, 4, 5, 6, time.Local),
			}
		})

		JustBeforeEach(func() {
			err := storage.SaveLogRecord("apptest", &logRecord)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should save log record", func() {
			cf := storage.cfmap["apptest"]
			Expect(cf).NotTo(BeNil())

			resp, err := storage.db.GetCF(
				gorocksdb.NewDefaultReadOptions(), cf, lastKey(cf),
			)
			Expect(err).To(BeNil())

			parsedRecord := decodeLogRecord(resp.Data())
			Expect(parsedRecord.Message).To(Equal(logRecord.Message))
			Expect(parsedRecord.Level).To(Equal(logRecord.Level))
			Expect(parsedRecord.Tags).To(Equal(logRecord.Tags))
			Expect(parsedRecord.CreatedAt).To(
				BeTemporally("~", logRecord.CreatedAt, time.Millisecond),
			)
		})

		Context("when CreatedAt of log record is zero", func() {
			BeforeEach(func() {
				logRecord.CreatedAt = time.Time{}
			})

			It("should set CreatedAt to log record", func() {
				Expect(logRecord.CreatedAt).To(
					BeTemporally("~", time.Now(), time.Second),
				)
			})
		})
	})

	Describe("loadLogRecords", func() {
		generateLogRecord := func(application, message string, level int, tags ...string) (logRecord LogRecord) {
			logRecord = LogRecord{
				Message: message,
				Level:   level,
				Tags:    tags,
			}
			Expect(
				storage.SaveLogRecord(application, &logRecord),
			).To(Succeed())

			time.Sleep(time.Millisecond)

			return
		}

		It("should return return log records filtered by app level and time", func() {
			logRecords := LogRecords{
				generateLogRecord("testapp1", "Message 1", 5),
				generateLogRecord("testapp2", "Message 2", 5),
				generateLogRecord("testapp1", "Message 3", 1),
				generateLogRecord("testapp1", "Message 4", 2, "tag1", "tag2"),
				generateLogRecord("testapp1", "Message 5", 5),
				generateLogRecord("testapp1", "Message 6", 5),
			}

			loadedLogRecords, err := storage.LoadLogRecords(
				"testapp1", 2, []string{},
				logRecords[1].CreatedAt, logRecords[4].CreatedAt, 1,
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(loadedLogRecords).To(HaveLen(2))

			Expect(loadedLogRecords[0].Level).To(Equal(logRecords[3].Level))
			Expect(loadedLogRecords[0].CreatedAt).To(
				BeTemporally("~", logRecords[3].CreatedAt, time.Millisecond),
			)
			Expect(loadedLogRecords[0].Message).To(Equal(logRecords[3].Message))
			Expect(loadedLogRecords[0].Tags).To(ConsistOf(logRecords[3].Tags))

			Expect(loadedLogRecords[1].Level).To(Equal(logRecords[4].Level))
			Expect(loadedLogRecords[1].CreatedAt).To(
				BeTemporally("~", logRecords[4].CreatedAt, time.Millisecond),
			)
			Expect(loadedLogRecords[1].Message).To(Equal(logRecords[4].Message))
			Expect(loadedLogRecords[1].Tags).To(BeEmpty())
		})

		Context("with tags", func() {
			It("should also filter log records by tags", func() {
				logRecords := LogRecords{
					generateLogRecord("testapp1", "Message 1", 5, "tag1", "tag2", "tag3"),
					generateLogRecord("testapp1", "Message 2", 5, "tag1", "tag2"),
					generateLogRecord("testapp1", "Message 3", 5, "tag2", "tag3"),
				}

				loadedLogRecords, err := storage.LoadLogRecords(
					"testapp1", 2, []string{"tag1", "tag2"},
					logRecords[0].CreatedAt, logRecords[2].CreatedAt, 1,
				)

				Expect(err).NotTo(HaveOccurred())

				Expect(loadedLogRecords).To(HaveLen(2))
				Expect(loadedLogRecords[0].Message).To(Equal(logRecords[0].Message))
				Expect(loadedLogRecords[1].Message).To(Equal(logRecords[1].Message))
			})
		})

		Context("with pagination", func() {
			It("should paginate results", func() {
				logRecords := make(LogRecords, 110)
				for i := 0; i < 110; i++ {
					logRecords[i] = generateLogRecord(
						"testapp1", fmt.Sprintf("Message%v", i), 5,
					)
				}

				loadedLogRecords, err := storage.LoadLogRecords(
					"testapp1", 2, []string{},
					logRecords[0].CreatedAt, logRecords[109].CreatedAt, 1,
				)

				Expect(err).NotTo(HaveOccurred())

				Expect(loadedLogRecords).To(HaveLen(100))
				Expect(loadedLogRecords[0].Message).To(Equal(logRecords[0].Message))
				Expect(loadedLogRecords[99].Message).To(Equal(logRecords[99].Message))

				loadedLogRecords, err = storage.LoadLogRecords(
					"testapp1", 2, []string{},
					logRecords[0].CreatedAt, logRecords[109].CreatedAt, 2,
				)

				Expect(err).NotTo(HaveOccurred())

				Expect(loadedLogRecords).To(HaveLen(10))
				Expect(loadedLogRecords[0].Message).To(Equal(logRecords[100].Message))
				Expect(loadedLogRecords[9].Message).To(Equal(logRecords[109].Message))
			})
		})
	})
})
