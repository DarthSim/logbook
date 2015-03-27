package main

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/boltdb/bolt"
)

var db *bolt.DB

const recordKeyFormat = "2006-02-01T15:04:05.000000000"

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

func tagKey(tag string) []byte {
	buf := bytes.NewBufferString("tag_")
	buf.WriteString(tag)
	return buf.Bytes()
}

func saveLogRecord(application string, logRecord *LogRecord) (err error) {
	err = db.Batch(func(tx *bolt.Tx) (err error) {
		logRecord.CreatedAt = time.Now()

		key := logRecord.CreatedAt.UTC().Format(recordKeyFormat)

		var valueBuf bytes.Buffer
		if err = gob.NewEncoder(&valueBuf).Encode(logRecord); err != nil {
			return
		}

		appBucket, err := tx.CreateBucketIfNotExists([]byte(application))
		if err != nil {
			return
		}

		recordBucket, err := appBucket.CreateBucket([]byte(key))
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

		if err = recordBucket.Put([]byte("record"), valueBuf.Bytes()); err != nil {
			return
		}

		return
	})
	return
}

func loadLogRecords(application string, lvl int, tags []string, startTime time.Time, endTime time.Time, page int) (logRecords LogRecords, err error) {
	err = db.View(func(tx *bolt.Tx) (err error) {
		appBucket := tx.Bucket([]byte(application))
		if appBucket == nil {
			return
		}

		cursor := appBucket.Cursor()

		keyStart := []byte(startTime.UTC().Format(recordKeyFormat))
		keyEnd := []byte(endTime.UTC().Format(recordKeyFormat))

		logRecords = LogRecords{}
		offset := (page - 1) * config.Pagination.PerPage

		for key, _ := cursor.Seek(keyStart); key != nil && bytes.Compare(key, keyEnd) <= 0; key, _ = cursor.Next() {
			recordBucket := appBucket.Bucket(key)
			if recordBucket == nil {
				// just for sure
				continue
			}

			recordLvl := recordBucket.Get([]byte("level"))
			if recordLvl == nil || recordLvl[0] < byte(lvl) {
				continue
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

			dec := gob.NewDecoder(bytes.NewBuffer(record))
			if err = dec.Decode(&logRecord); err != nil {
				return
			}

			logRecords = append(logRecords, logRecord)

			if len(logRecords) == config.Pagination.PerPage {
				break
			}
		}

		return
	})
	return
}
