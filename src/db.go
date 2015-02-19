package main

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"

	"github.com/coopernurse/gorp"
	_ "github.com/mattn/go-sqlite3"
)

var (
	dbmap         *gorp.DbMap
	dbInsertMutex sync.Mutex
)

// DB tools ========================================================================================

func initDBSchema() {
	dbmap.Exec("PRAGMA journal_mode=WAL;")

	dbmap.AddTableWithName(LogRecord{}, "log_records").SetKeys(true, "Id")
	dbmap.AddTableWithName(Application{}, "applications").SetKeys(true, "Id")
	dbmap.AddTableWithName(Tag{}, "tags").SetKeys(true, "Id")
	dbmap.AddTableWithName(LogRecordTag{}, "log_records_tags")

	err := dbmap.CreateTablesIfNotExists()
	checkErr(err, "Unable to create DB schema")

	dbmap.Exec("CREATE INDEX IF NOT EXISTS log_records_application_id_ind ON log_records (application_id)")
	dbmap.Exec("CREATE INDEX IF NOT EXISTS log_records_level_ind ON log_records (level)")
	dbmap.Exec("CREATE INDEX IF NOT EXISTS log_records_created_at_ind ON log_records (created_at)")

	dbmap.Exec("CREATE UNIQUE INDEX IF NOT EXISTS applications_name_ind ON applications (name)")

	dbmap.Exec("CREATE UNIQUE INDEX IF NOT EXISTS tags_name_ind ON tags (name)")

	dbmap.Exec("CREATE INDEX IF NOT EXISTS log_records_tags_log_record_id_ind ON log_records_tags (log_record_id)")
	dbmap.Exec("CREATE INDEX IF NOT EXISTS log_records_tags_tag_id_ind ON log_records_tags (tag_id)")
}

func initDB() {
	connectionString := fmt.Sprintf("file:%v?cache=shared&mode=rwc", absPathToFile(config.Database.Path))

	db, err := sql.Open("sqlite3", connectionString)
	checkErr(err, "sql.Open failed")

	dbmap = &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}

	dbmap.Db.SetMaxIdleConns(config.Database.MaxIdleConnections)

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

func dbSafeInsert(obj interface{}) error {
	dbInsertMutex.Lock()
	defer dbInsertMutex.Unlock()

	return dbmap.Insert(obj)
}

// end of DB tools
