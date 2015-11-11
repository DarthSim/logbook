package main

import (
	"fmt"
	"time"

	"github.com/boltdb/bolt"
	"gopkg.in/mgo.v2/bson"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Models", func() {
	decodeLogRecord := func(data []byte) (logRecord LogRecord) {
		err := bson.Unmarshal(data, &logRecord)
		Expect(err).NotTo(HaveOccurred())
		return
	}

	lastKey := func(bucket *bolt.Bucket) []byte {
		key, _ := bucket.Cursor().Last()
		return key
	}

	Describe("saveLogRecord", func() {
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
			err := saveLogRecord("apptest", &logRecord)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should save log record", func() {
			db.View(func(tx *bolt.Tx) (err error) {
				appBucket := tx.Bucket([]byte("apptest"))
				Expect(appBucket).NotTo(BeNil())

				recordBucket := appBucket.Bucket(lastKey(appBucket))
				Expect(recordBucket).NotTo(BeNil())

				Expect(recordBucket.Get([]byte("level"))).To(ConsistOf(byte(3)))
				Expect(recordBucket.Get([]byte("tag_tag1"))).To(ConsistOf(byte(1)))
				Expect(recordBucket.Get([]byte("tag_tag2"))).To(ConsistOf(byte(1)))

				parsedRecord := decodeLogRecord(
					recordBucket.Get([]byte("record")),
				)
				Expect(parsedRecord.Message).To(Equal(logRecord.Message))
				Expect(parsedRecord.Level).To(Equal(logRecord.Level))
				Expect(parsedRecord.Tags).To(Equal(logRecord.Tags))
				Expect(parsedRecord.CreatedAt.Truncate(time.Millisecond)).To(
					Equal(logRecord.CreatedAt.Truncate(time.Millisecond)),
				)

				return nil
			})
		})

		Context("when used many times", func() {
			It("should save log record many times", func() {
				err := saveLogRecord("apptest", &logRecord)
				Expect(err).NotTo(HaveOccurred())

				db.View(func(tx *bolt.Tx) (err error) {
					appBucket := tx.Bucket([]byte("apptest"))
					Expect(appBucket).NotTo(BeNil())

					recordBucket := appBucket.Bucket(lastKey(appBucket))
					Expect(recordBucket).NotTo(BeNil())

					Expect(recordBucket.Get([]byte("level"))).To(ConsistOf(byte(3)))
					Expect(recordBucket.Get([]byte("tag_tag1"))).To(ConsistOf(byte(1)))
					Expect(recordBucket.Get([]byte("tag_tag2"))).To(ConsistOf(byte(1)))

					parsedRecord := decodeLogRecord(
						recordBucket.Get([]byte("record")),
					)
					Expect(parsedRecord.Message).To(Equal(logRecord.Message))
					Expect(parsedRecord.Level).To(Equal(logRecord.Level))
					Expect(parsedRecord.Tags).To(Equal(logRecord.Tags))
					Expect(parsedRecord.CreatedAt.Truncate(time.Millisecond)).To(
						Equal(logRecord.CreatedAt.Truncate(time.Millisecond)),
					)

					return nil
				})
			})
		})

		Context("when CreatedAt of log record is zero", func() {
			BeforeEach(func() {
				logRecord.CreatedAt = time.Time{}
			})

			It("should set CreatedAt to log record", func() {
				Expect(logRecord.CreatedAt).NotTo(BeNil())
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
				saveLogRecord(application, &logRecord),
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

			loadedLogRecords, err := loadLogRecords("testapp1", 2, []string{},
				logRecords[1].CreatedAt, logRecords[4].CreatedAt, 1)

			Expect(err).NotTo(HaveOccurred())
			Expect(loadedLogRecords).To(HaveLen(2))

			Expect(loadedLogRecords[0].Level).To(Equal(logRecords[3].Level))
			Expect(loadedLogRecords[0].CreatedAt.Truncate(time.Millisecond)).To(
				Equal(logRecords[3].CreatedAt.Truncate(time.Millisecond)),
			)
			Expect(loadedLogRecords[0].Message).To(Equal(logRecords[3].Message))
			Expect(loadedLogRecords[0].Tags).To(Equal(logRecords[3].Tags))

			Expect(loadedLogRecords[1].Level).To(Equal(logRecords[4].Level))
			Expect(loadedLogRecords[1].CreatedAt.Truncate(time.Millisecond)).To(
				Equal(logRecords[4].CreatedAt.Truncate(time.Millisecond)),
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

				loadedLogRecords, err := loadLogRecords("testapp1", 2, []string{"tag1", "tag2"},
					logRecords[0].CreatedAt, logRecords[2].CreatedAt, 1)

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

				loadedLogRecords, err := loadLogRecords("testapp1", 2, []string{},
					logRecords[0].CreatedAt, logRecords[109].CreatedAt, 1)

				Expect(err).NotTo(HaveOccurred())

				Expect(loadedLogRecords).To(HaveLen(100))
				Expect(loadedLogRecords[0].Message).To(Equal(logRecords[0].Message))
				Expect(loadedLogRecords[99].Message).To(Equal(logRecords[99].Message))

				loadedLogRecords, err = loadLogRecords("testapp1", 2, []string{},
					logRecords[0].CreatedAt, logRecords[109].CreatedAt, 2)

				Expect(err).NotTo(HaveOccurred())

				Expect(loadedLogRecords).To(HaveLen(10))
				Expect(loadedLogRecords[0].Message).To(Equal(logRecords[100].Message))
				Expect(loadedLogRecords[9].Message).To(Equal(logRecords[109].Message))
			})
		})
	})
})
