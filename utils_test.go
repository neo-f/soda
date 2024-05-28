package soda

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestToIntE(t *testing.T) {
	convey.Convey("Given a string that can be converted to an integer", t, func() {
		convey.Convey("When the string is '123'", func() {
			result, err := toIntE("123")

			convey.Convey("Then the result should be 123", func() {
				convey.So(result, convey.ShouldEqual, 123)
			})

			convey.Convey("And there should be no error", func() {
				convey.So(err, convey.ShouldBeNil)
			})
		})

		convey.Convey("When the string is '-123'", func() {
			result, err := toIntE("-123")

			convey.Convey("Then the result should be -123", func() {
				convey.So(result, convey.ShouldEqual, -123)
			})

			convey.Convey("And there should be no error", func() {
				convey.So(err, convey.ShouldBeNil)
			})
		})
	})

	convey.Convey("Given a string that cannot be converted to an integer", t, func() {
		convey.Convey("When the string is 'abc'", func() {
			_, err := toIntE("abc")

			convey.Convey("Then there should be an error", func() {
				convey.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestToUint64E(t *testing.T) {
	convey.Convey("Given a string that can be converted to a uint64", t, func() {
		convey.Convey("When the string is '123'", func() {
			result, err := toUint64E("123")

			convey.Convey("Then the result should be 123", func() {
				convey.So(result, convey.ShouldEqual, 123)
			})

			convey.Convey("And there should be no error", func() {
				convey.So(err, convey.ShouldBeNil)
			})
		})
	})

	convey.Convey("Given a string that cannot be converted to a uint64", t, func() {
		convey.Convey("When the string is '-123'", func() {
			_, err := toUint64E("-123")

			convey.Convey("Then there should be an error", func() {
				convey.So(err, convey.ShouldNotBeNil)
			})
		})

		convey.Convey("When the string is 'abc'", func() {
			_, err := toUint64E("abc")

			convey.Convey("Then there should be an error", func() {
				convey.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}
