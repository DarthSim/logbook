package main

import (
	"os"
	"testing"

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

	initLogger()
})

var _ = AfterSuite(func() {
	closeLogger()
	os.Remove(absPathToFile(config.Log.Path))
})

func TestLogbook(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Logbook Suite")
}
