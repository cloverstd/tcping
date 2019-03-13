package ping


import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFormatIP(t *testing.T) {

	Convey("IP", t, func() {
		Convey("for v4 success", func() {
			rc := FormatIP("192.168.0.1")
			So(rc, ShouldEqual, "192.168.0.1")
		})

		Convey("for v4 failure", func() {
			rc := FormatIP("192.0.1")
			So(rc, ShouldEqual, "")
		})

		Convey("for v4 format", func() {
			rc := FormatIP("[192.0.1.1] ")
			So(rc, ShouldEqual, "192.0.1.1")
		})

		Convey("for v6 success", func() {
			rc := FormatIP("[2002:ac1f:91c5:1::bd59]")
			So(rc, ShouldEqual, "[2002:ac1f:91c5:1::bd59]")
		})

		Convey("for v6 failure", func() {
			rc := FormatIP("2002:ac1f:91c5:1:")
			So(rc, ShouldEqual, "")
		})

		Convey("for v6 format", func() {
			rc := FormatIP("2002:ac1f:91c5:1::bd59 ")
			So(rc, ShouldEqual, "[2002:ac1f:91c5:1::bd59]")
		})
	})
}
