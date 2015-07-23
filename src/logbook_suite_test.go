package main

import (
	"os"
	"testing"

	"github.com/boltdb/bolt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = BeforeSuite(func() {
	config = Config{}

	config.Database.Path = "test.db"

	config.Auth.User = "test"
	config.Auth.Password = "test"

	config.Pagination.PerPage = 100

	initDB()
})

var _ = AfterSuite(func() {
	closeDB()
	os.Remove(absPathToFile(config.Database.Path))
})

var _ = BeforeEach(func() {
	db.Update(func(tx *bolt.Tx) (err error) {
		err = tx.ForEach(func(name []byte, b *bolt.Bucket) (err error) {
			err = tx.DeleteBucket(name)
			return
		})
		return
	})
})

func TestLogbook(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Logbook Suite")
}
