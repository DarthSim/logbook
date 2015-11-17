package main

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Utils", func() {
	Describe("absPathToFile", func() {
		It("should return provided absolute path as is", func() {
			Expect(absPathToFile("/lorem/ipsum/dolor")).To(
				Equal("/lorem/ipsum/dolor"),
			)
		})

		It("should expand provided relative path", func() {
			Expect(absPathToFile("./lorem/ipsum/dolor")).To(
				Equal(appPath() + "/lorem/ipsum/dolor"),
			)
		})
	})

	Describe("checkTimeFormat", func() {
		It("should return true for valid datetime format", func() {
			Expect(checkTimeFormat("2014-08-08T01:02:03")).To(BeTrue())
			Expect(checkTimeFormat("2014-08-08T01:02:03+06:00")).To(BeTrue())
			Expect(checkTimeFormat("2014-08-08T01:02:03.123")).To(BeTrue())
			Expect(checkTimeFormat("2014-08-08T01:02:03.123+06:00")).To(BeTrue())
		})

		It("should return false for invalid datetime format", func() {
			Expect(checkTimeFormat("2014-08-08T01:02:033")).To(BeFalse())
		})
	})

	Describe("checkDateTimeFormat", func() {
		It("should return true for valid date format", func() {
			Expect(checkDateTimeFormat("2014-08-08")).To(BeTrue())
		})

		It("should return true for valid datetime format", func() {
			Expect(checkDateTimeFormat("2014-08-08T01:02:03")).To(BeTrue())
			Expect(checkDateTimeFormat("2014-08-08T01:02:03+06:00")).To(BeTrue())
			Expect(checkDateTimeFormat("2014-08-08T01:02:03.123")).To(BeTrue())
			Expect(checkDateTimeFormat("2014-08-08T01:02:03.123+06:00")).To(BeTrue())
		})

		It("should return false for invalid date format", func() {
			Expect(checkDateTimeFormat("2014-08-088")).To(BeFalse())
		})

		It("should return false for invalid datetime format", func() {
			Expect(checkDateTimeFormat("2014-08-08T01:02:033")).To(BeFalse())
		})
	})

	Describe("parseTime", func() {
		It("should parse time", func() {
			result, err := parseTime("2014-09-08T11:12:13.321")
			Expect(result).To(
				Equal(time.Date(2014, 9, 8, 11, 12, 13, 321000000, time.Local)),
			)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when provided string has invalid format", func() {
			It("should return error", func() {
				_, err := parseTime("2014-09-08T11:12:134.321")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("parseDateTime", func() {
		It("should parse time", func() {
			result, err := parseDateTime("2014-09-08T11:12:13.321", false)
			Expect(result).To(
				Equal(time.Date(2014, 9, 8, 11, 12, 13, 321000000, time.Local)),
			)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should parse date and set time to the beginning of day", func() {
			result, err := parseDateTime("2014-09-08", false)
			Expect(result).To(Equal(time.Date(2014, 9, 8, 0, 0, 0, 0, time.Local)))
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when clockToEnd is true", func() {
			It("should parse date and set time to the end of day", func() {
				result, err := parseDateTime("2014-09-08", true)
				Expect(result).To(Equal(time.Date(2014, 9, 8, 23, 59, 59, 999999999, time.Local)))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when provided string has invalid format", func() {
			It("should return error", func() {
				_, err := parseDateTime("2014-09-088", false)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("uniqStrings", func() {
		It("should remove dublicated items from array", func() {
			input := []string{"fff", "fff"}
			Expect(uniqStrings(input)).To(Equal([]string{"fff"}))
		})

		It("should return array with all elements from provided one", func() {
			input := []string{"fff", "fff2"}
			Expect(uniqStrings(input)).To(HaveLen(len(input)))
			Expect(uniqStrings(input)).To(ConsistOf(input))
		})
	})
})
