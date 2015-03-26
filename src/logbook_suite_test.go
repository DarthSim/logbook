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

	config.Database.Path = "test.sqlite"
	config.Database.LockTimeout = 1
	config.Database.RetryDelay = 10
	config.Database.MaxOpenConnections = 5
	config.Database.MaxIdleConnections = 5

	config.Log.Path = "test.log"
	config.Log.LogDatabase = false

	config.Auth.User = "test"
	config.Auth.Password = "test"

	config.Pagination.PerPage = 100

	initLogger()

	initDB()
})

var _ = AfterSuite(func() {
	closeDB()
	os.Remove(absPathToFile(config.Database.Path))

	closeLogger()
	os.Remove(absPathToFile(config.Log.Path))
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
