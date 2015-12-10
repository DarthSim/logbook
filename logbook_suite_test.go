package main

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = BeforeSuite(func() {
	config = Config{}

	config.Database.Path = "/tmp/logbook_test_db"

	config.Auth.User = "test"
	config.Auth.Password = "test"

	config.Pagination.PerPage = 100

	OpenStorage()
})

var _ = AfterSuite(func() {
	storage.Close()
	os.RemoveAll(absPathToFile(config.Database.Path))
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
