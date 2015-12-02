package main

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LogRecord", func() {
	It("can be encoded and decoded", func() {
		record1 := LogRecord{
			Message:   "Lorem ipsum dolor",
			Level:     3,
			Tags:      []string{"tag321", "tag123"},
			CreatedAt: time.Date(2015, 1, 2, 3, 4, 5, 123000000, time.Local),
		}

		var record2 LogRecord

		data := record1.Encode()
		err := record2.Decode(data)

		Expect(err).To(BeNil())
		Expect(record2).To(Equal(record1))
	})
})
