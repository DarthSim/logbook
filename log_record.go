package main

import (
	"bytes"
	"errors"
	"strings"
	"time"
)

var invalidLogRecordFormat = errors.New("Invalid log record format")

type LogRecord struct {
	Message   string    `json:"message"`
	Level     int       `json:"level"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
}

type LogRecords []LogRecord

func (r *LogRecord) Encode() []byte {
	buf := bytes.NewBuffer([]byte{byte(r.Level)})
	buf.WriteString(strings.Join(r.Tags, ","))
	buf.WriteByte('\n')
	buf.WriteString(r.CreatedAt.Format(dbTimeFormat))
	buf.WriteByte('\n')
	buf.WriteString(r.Message)
	return buf.Bytes()
}

func (r *LogRecord) Decode(src []byte) error {
	buf := bytes.NewBuffer(src)

	lvl, err := buf.ReadByte()
	if err != nil {
		return invalidLogRecordFormat
	}
	r.Level = int(lvl)

	tags, err := buf.ReadBytes('\n')
	if err != nil {
		return invalidLogRecordFormat
	}
	if len(tags) > 1 {
		r.Tags = strings.Split(string(tags[:len(tags)-1]), ",")
	} else {
		r.Tags = []string{}
	}

	created, err := buf.ReadBytes('\n')
	if err != nil {
		return invalidLogRecordFormat
	}
	r.CreatedAt, err = time.Parse(
		dbTimeFormat,
		string(created[:len(created)-1]),
	)
	if err != nil {
		return invalidLogRecordFormat
	}

	r.Message = buf.String()

	return nil
}
