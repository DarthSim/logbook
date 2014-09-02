package main

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/coopernurse/gorp"
	_ "github.com/mattn/go-sqlite3"
)

var dbmap *gorp.DbMap

// DB tools ========================================================================================

func initDBSchema() {
	dbmap.AddTableWithName(LogRecord{}, "log_records").SetKeys(true, "Id")
	dbmap.AddTableWithName(Application{}, "applications").SetKeys(true, "Id")
	dbmap.AddTableWithName(Tag{}, "tags").SetKeys(true, "Id")
	dbmap.AddTableWithName(LogRecordTag{}, "log_records_tags")

	err := dbmap.CreateTablesIfNotExists()
	checkErr(err, "Unable to create DB schema")

	dbmap.Exec("CREATE INDEX IF NOT EXISTS log_records_application_id_ind ON log_records (application_id)")
	dbmap.Exec("CREATE INDEX IF NOT EXISTS log_records_level_ind ON log_records (level)")
	dbmap.Exec("CREATE INDEX IF NOT EXISTS log_records_created_at_ind ON log_records (created_at)")

	dbmap.Exec("CREATE INDEX IF NOT EXISTS applications_name_ind ON applications (name)")

	dbmap.Exec("CREATE INDEX IF NOT EXISTS tags_name_ind ON tags (name)")

	dbmap.Exec("CREATE INDEX IF NOT EXISTS log_records_tags_log_record_id_ind ON log_records_tags (log_record_id)")
	dbmap.Exec("CREATE INDEX IF NOT EXISTS log_records_tags_tag_id_ind ON log_records_tags (tag_id)")
}

func initDB() {
	connectionString := fmt.Sprintf("file:%v?cache=shared&mode=rwc", config.Database.Path)

	db, err := sql.Open("sqlite3", connectionString)
	checkErr(err, "sql.Open failed")

	dbmap = &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}

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

// end of DB tools
