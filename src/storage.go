package main

import (
	"bytes"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
	"gopkg.in/mgo.v2/bson"
)

var db *bolt.DB

type LogRecord struct {
	Message   string    `json:"message"`
	Level     int       `json:"level"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
}

type LogRecords []LogRecord

func initDB() {
	var err error

	db, err = bolt.Open(absPathToFile(config.Database.Path), 0600, nil)
	checkErr(err, "bolt.Open failed")
}

func closeDB() {
	db.Close()
}

func recordKey(createdAt time.Time, suffix string) []byte {
	buf := bytes.NewBufferString(
		createdAt.UTC().Format("2006-02-01T15:04:05.000"),
	)
	buf.WriteString("_")
	buf.WriteString(suffix)
	return buf.Bytes()
}

func tagKey(tag string) []byte {
	buf := bytes.NewBufferString("tag_")
	buf.WriteString(tag)
	return buf.Bytes()
}

func saveLogRecord(application string, logRecord *LogRecord) (err error) {
	err = db.Batch(func(tx *bolt.Tx) (err error) {
		if logRecord.CreatedAt.IsZero() {
			logRecord.CreatedAt = time.Now()
		}

		data, err := bson.Marshal(&logRecord)
		if err != nil {
			return
		}

		appBucket, err := tx.CreateBucketIfNotExists([]byte(application))
		if err != nil {
			return
		}

		id, _ := appBucket.NextSequence()
		key := recordKey(logRecord.CreatedAt, strconv.Itoa(int(id)))
		recordBucket, err := appBucket.CreateBucket(key)
		if err != nil {
			return
		}

		if err = recordBucket.Put([]byte("level"), []byte{byte(logRecord.Level)}); err != nil {
			return
		}

		for _, tag := range logRecord.Tags {
			if err = recordBucket.Put(tagKey(tag), []byte{1}); err != nil {
				return
			}
		}

		if err = recordBucket.Put([]byte("record"), data); err != nil {
			return
		}

		return
	})
	return
}

func loadLogRecords(application string, lvl int, tags []string, startTime time.Time, endTime time.Time, page int) (logRecords LogRecords, err error) {
	logRecords = LogRecords{}

	err = db.View(func(tx *bolt.Tx) (err error) {
		appBucket := tx.Bucket([]byte(application))
		if appBucket == nil {
			return
		}

		cursor := appBucket.Cursor()

		keyStart := recordKey(startTime, "")
		keyEnd := recordKey(endTime, "_")

		offset := (page - 1) * config.Pagination.PerPage

		for key, _ := cursor.Seek(keyStart); key != nil && bytes.Compare(key, keyEnd) <= 0; key, _ = cursor.Next() {
			recordBucket := appBucket.Bucket(key)
			if recordBucket == nil {
				// just for sure
				continue
			}

			if lvl > 0 {
				recordLvl := recordBucket.Get([]byte("level"))
				if recordLvl == nil || recordLvl[0] < byte(lvl) {
					continue
				}
			}

			tagMissed := false
			for _, tag := range tags {
				if recordBucket.Get(tagKey(tag)) == nil {
					tagMissed = true
					break
				}
			}
			if tagMissed {
				continue
			}

			if offset > 0 {
				offset--
				continue
			}

			var logRecord LogRecord

			record := recordBucket.Get([]byte("record"))
			if record == nil {
				continue
			}

			if err = bson.Unmarshal(record, &logRecord); err != nil {
				return
			}

			logRecords = append(logRecords, logRecord)

			if len(logRecords) >= config.Pagination.PerPage {
				break
			}
		}

		return
	})
	return
}
