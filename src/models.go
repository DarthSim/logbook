package main

import (
	"database/sql"
	"strconv"
	"strings"
	"time"
)

// Models ==========================================================================================

type LogRecord struct {
	Id            int64      `db:"id"`
	ApplicationId int64      `db:"application_id"`
	Level         int        `db:"level"`
	CreatedAt     *time.Time `db:"created_at"`
	Message       string     `db:"message"`

	Application Application `db:"-"`
	Tags        []Tag       `db:"-"`
}

type Application struct {
	Id   int64  `db:"id"`
	Name string `db:"name"`
}

type Tag struct {
	Id   int64  `db:"id"`
	Name string `db:"name"`
}

type LogRecordTag struct {
	LogRecordId int64 `db:"log_record_id"`
	TagId       int64 `db:"tag_id"`
}

// end of Models

// Application =====================================================================================

func findOrCreateApplication(name string) (Application, error) {
	var application Application

	err := dbmap.SelectOne(&application, "SELECT * FROM applications WHERE name = $1", name)

	if err == sql.ErrNoRows {
		application.Name = name

		err := dbmap.Insert(&application)
		if err != nil {
			return Application{}, err
		}
	} else if err != nil {
		return Application{}, err
	}

	return application, nil
}

// end of Application

// Tag =============================================================================================

func tagNamesInOptions(names []string) string {
	escapedNames := make([]string, len(names))

	for i, name := range names {
		escapedNames[i] = dbEscapeString(name)
	}

	return strings.Join(escapedNames, "','")
}

func findTagIds(names []string) ([]string, error) {
	var tagIds []string

	joinedNames := tagNamesInOptions(names)
	_, err := dbmap.Select(&tagIds, "SELECT id FROM tags WHERE name IN ('"+joinedNames+"')")

	return tagIds, err
}

func findTags(names []string) ([]Tag, error) {
	var tags []Tag

	joinedNames := tagNamesInOptions(names)
	_, err := dbmap.Select(&tags, "SELECT * FROM tags WHERE name IN ('"+joinedNames+"')")

	return tags, err
}

func createLogRecordTag(logRecord *LogRecord, tag *Tag) (LogRecordTag, error) {
	var logRecordTag = LogRecordTag{
		LogRecordId: logRecord.Id,
		TagId:       tag.Id,
	}

	err := dbmap.Insert(&logRecordTag)
	if err != nil {
		return LogRecordTag{}, err
	}

	return logRecordTag, nil
}

func addTagsToLogRecord(logRecord *LogRecord, tagNames []string) error {
	existingTags, err := findTags(tagNames)
	if err != nil {
		return err
	}

	newTagsCount := len(tagNames) - len(existingTags)

	if newTagsCount > 0 {
		newTags := make([]Tag, newTagsCount)

		var exists bool

		i := 0

		for _, tagName := range tagNames {
			exists = false

			for _, tag := range existingTags {
				if tag.Name == tagName {
					exists = true
					break
				}
			}

			if !exists {
				newTags[i] = Tag{
					Name: tagName,
				}

				err = dbmap.Insert(&newTags[i])
				if err != nil {
					return err
				}

				i++
			}
		}

		existingTags = append(existingTags, newTags...)
	}

	for _, tag := range existingTags {
		_, err := createLogRecordTag(logRecord, &tag)
		if err != nil {
			return err
		}
	}

	logRecord.Tags = existingTags

	return nil
}

func loadTagsOfLogRecords(logRecords []LogRecord) error {
	var tags []Tag
	var logRecordsTags []LogRecordTag

	logRecordIds := make([]string, len(logRecords))
	for i, logRecord := range logRecords {
		logRecordIds[i] = strconv.FormatInt(logRecord.Id, 10)
	}

	_, err := dbmap.Select(&logRecordsTags, "SELECT * FROM log_records_tags WHERE log_record_id IN ("+strings.Join(logRecordIds, ",")+")")
	if err != nil {
		return err
	}

	tagIds := []string{}
	for _, logRecordTag := range logRecordsTags {
		tagId := strconv.FormatInt(logRecordTag.TagId, 10)
		if indexOfString(tagIds, tagId) == -1 {
			tagIds = append(tagIds, tagId)
		}
	}

	_, err = dbmap.Select(&tags, "SELECT * FROM tags WHERE id IN ("+strings.Join(tagIds, ",")+")")

	for i, logRecord := range logRecords {
		logRecords[i].Tags = []Tag{}
		for _, logRecordTag := range logRecordsTags {
			if logRecord.Id == logRecordTag.LogRecordId {
				for _, tag := range tags {
					if tag.Id == logRecordTag.TagId {
						logRecords[i].Tags = append(logRecords[i].Tags, tag)
					}
				}
			}
		}
	}

	return nil
}

// end of Tag

// LogRecord =======================================================================================

func createLogRecord(appName string, msg string, lvl int, tagNames []string) (LogRecord, error) {
	now := time.Now()

	logRecord := LogRecord{
		Level:     lvl,
		CreatedAt: &now,
		Message:   msg,
	}

	application, err := findOrCreateApplication(appName)
	if err != nil {
		return LogRecord{}, err
	}

	logRecord.ApplicationId = application.Id
	logRecord.Application = application

	err = dbmap.Insert(&logRecord)
	if err != nil {
		return LogRecord{}, err
	}

	if len(tagNames) > 0 {
		uniqTagNames := uniqStrings(tagNames)

		err := addTagsToLogRecord(&logRecord, uniqTagNames)
		if err != nil {
			return LogRecord{}, err
		}
	}

	return logRecord, nil
}

func findLogRecords(appName string, lvl int, tagNames []string, startTime *time.Time, endTime *time.Time) ([]LogRecord, error) {
	application, err := findOrCreateApplication(appName)
	if err != nil {
		return nil, err
	}

	var (
		join    string
		groupBy string
		having  string
		where   string

		tagsCount int
	)

	if len(tagNames) > 0 {
		uniqTagNames := uniqStrings(tagNames)

		tagIds, err := findTagIds(uniqTagNames)
		if err != nil {
			return nil, err
		}

		if len(tagIds) < len(uniqTagNames) {
			// Some tags weren't found
			return []LogRecord{}, nil
		}

		join = "INNER JOIN log_records_tags ON log_records_tags.log_record_id = log_records.id"
		where = "AND log_records_tags.tag_id IN (" + strings.Join(tagIds, ",") + ")"
		groupBy = "GROUP BY log_records.id"
		having = "HAVING COUNT(log_records_tags.log_record_id) = :tags_count"

		tagsCount = len(tagIds)
	}

	var logRecords []LogRecord

	_, err = dbmap.Select(&logRecords, `
	  SELECT DISTINCT log_records.*
	  FROM log_records `+join+`
	  WHERE log_records.application_id = :application_id
	    AND log_records.level >= :level
	    AND log_records.created_at >= :start_time
	    AND log_records.created_at <= :end_time
	  `+where+`
	  `+groupBy+`
	  `+having+`
  `, map[string]interface{}{
		"application_id": application.Id,
		"level":          lvl,
		"start_time":     startTime,
		"end_time":       endTime,
		"tags_count":     tagsCount,
	})

	if err != nil {
		return nil, err
	}

	for i, _ := range logRecords {
		logRecords[i].Application = application
	}

	err = loadTagsOfLogRecords(logRecords)
	if err != nil {
		return nil, err
	}

	return logRecords, nil
}

// end of LogRecord
