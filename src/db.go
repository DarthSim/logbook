package main

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-gorp/gorp"
	"github.com/mattn/go-sqlite3"
)

var dbmap *gorp.DbMap

// DB tools ========================================================================================

func initDBSchema() {
	dbmap.Exec("PRAGMA journal_mode = WAL;")
	dbmap.Exec("PRAGMA synchronous = NORMAL;")
	dbmap.Exec("PRAGMA encoding = \"UTF-8\";")

	dbmap.AddTableWithName(LogRecord{}, "log_records").SetKeys(true, "Id")
	dbmap.AddTableWithName(Application{}, "applications").SetKeys(true, "Id")
	dbmap.AddTableWithName(Tag{}, "tags").SetKeys(true, "Id")
	dbmap.AddTableWithName(LogRecordTag{}, "log_records_tags")

	err := dbmap.CreateTablesIfNotExists()
	checkErr(err, "Unable to create DB schema")

	_, err = dbmap.Exec(`
		CREATE INDEX IF NOT EXISTS log_records_application_full_ind ON log_records (application_id, level, created_at);
		CREATE UNIQUE INDEX IF NOT EXISTS applications_name_ind ON applications (name);
		CREATE UNIQUE INDEX IF NOT EXISTS tags_name_ind ON tags (name);
		CREATE UNIQUE INDEX IF NOT EXISTS log_records_tags_full_ind ON log_records_tags (log_record_id, tag_id);
	`)
	checkErr(err, "Unable to create DB schema")
}

func initDB() {
	connectionString := fmt.Sprintf(
		"file:%v?cache=shared&mode=rwc",
		absPathToFile(config.Database.Path),
	)

	db, err := sql.Open("sqlite3", connectionString)
	checkErr(err, "sql.Open failed")

	dbmap = &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}

	dbmap.Db.SetMaxIdleConns(config.Database.MaxIdleConnections)
	dbmap.Db.SetMaxOpenConns(config.Database.MaxOpenConnections)

	initDBSchema()

	if config.Log.LogDatabase {
		dbmap.TraceOn("[database]", logger)
	}
}

func closeDB() {
	dbmap.Db.Close()
}

func dbEscapeString(str string) string {
	return strings.Replace(str, "'", "''", -1)
}

func dbIsErrLocked(err error) bool {
	return err != nil &&
		reflect.TypeOf(err).String() == "sqlite3.Error" &&
		err.(sqlite3.Error).Code == sqlite3.ErrLocked
}

func dbSafeInsert(obj interface{}) (err error) {
	tries := int64(0)
	maxTries := config.Database.LockTimeout / config.Database.RetryDelay
	retryDelay := time.Millisecond * time.Duration(config.Database.RetryDelay)

	for {
		tries++
		err = dbmap.Insert(obj)
		if !dbIsErrLocked(err) || tries >= maxTries {
			break
		}
		time.Sleep(retryDelay)
	}

	return
}

// end of DB tools
