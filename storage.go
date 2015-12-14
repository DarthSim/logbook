package main

import (
	"bytes"
	"encoding/binary"
	"strconv"
	"sync"
	"time"

	"github.com/tecbot/gorocksdb"
)

const dbTimeFormat = "2006-01-02T15:04:05.000-07:00"

type Storage struct {
	db       *gorocksdb.DB
	cfmap    map[string]*gorocksdb.ColumnFamilyHandle
	seqMutex sync.Mutex
}

var storage Storage

func OpenStorage() {
	dbopts := gorocksdb.NewDefaultOptions()
	dbopts.SetCreateIfMissing(true)

	cfnames, err := gorocksdb.ListColumnFamilies(dbopts, config.DBPath)
	if err != nil {
		cfnames = []string{"default"}
	}

	cfopts := make([]*gorocksdb.Options, len(cfnames))
	for i := range cfopts {
		cfopts[i] = gorocksdb.NewDefaultOptions()
	}

	var cfhandles []*gorocksdb.ColumnFamilyHandle

	storage.db, cfhandles, err = gorocksdb.OpenDbColumnFamilies(
		dbopts, config.DBPath, cfnames, cfopts,
	)
	checkErr(err, "can't open RocksDB database")

	storage.cfmap = make(map[string]*gorocksdb.ColumnFamilyHandle)
	for i, name := range cfnames {
		storage.cfmap[name] = cfhandles[i]
	}
}

func (s *Storage) Close() {
	for _, h := range s.cfmap {
		h.Destroy()
	}
	s.db.Close()
}

func recordKey(createdAt time.Time, suffix string) []byte {
	buf := bytes.NewBufferString(
		createdAt.UTC().Format(dbTimeFormat),
	)
	buf.WriteString("_")
	buf.WriteString(suffix)
	return buf.Bytes()
}

func (s *Storage) nextSeq(cf *gorocksdb.ColumnFamilyHandle) (seq uint64, err error) {
	s.seqMutex.Lock()
	defer s.seqMutex.Unlock()

	resp, err := s.db.GetCF(
		gorocksdb.NewDefaultReadOptions(), cf, []byte("::seq::"),
	)
	if err != nil {
		return
	}

	if resp.Size() > 0 {
		seq, _ = binary.Uvarint(resp.Data())
		seq++
	}

	buf := make([]byte, binary.MaxVarintLen64)
	binary.PutUvarint(buf, seq)

	err = s.db.PutCF(
		gorocksdb.NewDefaultWriteOptions(), cf, []byte("::seq::"), buf,
	)

	return
}

func (s *Storage) getOrCreateCF(name string) (cf *gorocksdb.ColumnFamilyHandle, err error) {
	cf, ok := s.cfmap[name]
	if !ok {
		cf, err = s.db.CreateColumnFamily(gorocksdb.NewDefaultOptions(), name)
		s.cfmap[name] = cf
	}
	return
}

func (s *Storage) SaveLogRecord(application string, logRecord *LogRecord) (err error) {
	if logRecord.CreatedAt.IsZero() {
		logRecord.CreatedAt = time.Now()
	}

	cf, err := s.getOrCreateCF(application)
	if err != nil {
		return
	}

	id, err := s.nextSeq(cf)
	if err != nil {
		return
	}

	return s.db.PutCF(
		gorocksdb.NewDefaultWriteOptions(),
		cf,
		recordKey(logRecord.CreatedAt, strconv.FormatUint(id, 16)),
		logRecord.Encode(),
	)
}

func (s *Storage) LoadLogRecords(application string, lvl int, tags []string, startTime time.Time, endTime time.Time, page int) (logRecords LogRecords, err error) {
	keyStart := recordKey(startTime, "")
	keyEnd := recordKey(endTime, "_")

	offset := (page - 1) * config.RecordsPerPage

	records := make(LogRecords, config.RecordsPerPage)
	fetched := 0

	cf, err := s.getOrCreateCF(application)
	if err != nil {
		return
	}

	it := s.db.NewIteratorCF(gorocksdb.NewDefaultReadOptions(), cf)
	defer it.Close()

	var record LogRecord

	for it.Seek(keyStart); it.Valid() && bytes.Compare(it.Key().Data(), keyEnd) <= 0; it.Next() {
		if it.Value().Size() == 0 {
			// just for sure
			continue
		}

		err = record.Decode(it.Value().Data())
		if err != nil {
			return
		}

		if lvl > record.Level {
			continue
		}

		if !stringsContain(record.Tags, tags) {
			continue
		}

		if offset > 0 {
			offset--
			continue
		}

		records[fetched] = record

		fetched++
		if fetched == config.RecordsPerPage {
			break
		}
	}

	return records[:fetched], nil
}

func (s *Storage) appStats(application string) (stats string, err error) {
	cf, err := s.getOrCreateCF(application)
	if err == nil {
		stats = s.db.GetPropertyCF("rocksdb.stats", cf)
	}
	return
}
