package main

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = BeforeSuite(func() {
	config = Config{}

	config.DBPath = "/tmp/logbook_test_db"

	config.Username = "test"
	config.Password = "test"

	config.RecordsPerPage = 100

	OpenStorage()
})

var _ = AfterSuite(func() {
	storage.Close()
	os.RemoveAll(absPathToFile(config.DBPath))
})

var _ = BeforeEach(func() {
	for name, cf := range storage.cfmap {
		if name == "default" {
			continue
		}

		storage.db.DropColumnFamily(cf)
		cf.Destroy()
		delete(storage.cfmap, name)
	}
})

func TestLogbook(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Logbook Suite")
}
