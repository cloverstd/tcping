package ping

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFormatIP(t *testing.T) {

	Convey("IP", t, func() {
		Convey("for v4 success", func() {
			rc, _ := FormatIP("192.168.0.1")
			So(rc, ShouldEqual, "192.168.0.1")
		})

		Convey("for v4 failure", func() {
			rc, _ := FormatIP("192.0.1")
			So(rc, ShouldEqual, "")
		})

		Convey("for v4 format", func() {
			rc, _ := FormatIP("[192.0.1.1] ")
			So(rc, ShouldEqual, "192.0.1.1")
		})

		Convey("for v6 success", func() {
			rc, _ := FormatIP("[2002:ac1f:91c5:1::bd59]")
			So(rc, ShouldEqual, "[2002:ac1f:91c5:1::bd59]")
		})

		Convey("for v6 failure", func() {
			rc, _ := FormatIP("2002:ac1f:91c5:1:")
			So(rc, ShouldEqual, "")
		})

		Convey("for v6 format", func() {
			rc, _ := FormatIP("2002:ac1f:91c5:1::bd59 ")
			So(rc, ShouldEqual, "[2002:ac1f:91c5:1::bd59]")
		})
	})
}

func TestParseAddress(t *testing.T) {
	Convey("ParseAddress", t, func() {
		Convey("formats bare ipv6 host", func() {
			u, err := ParseAddress("2001:db8::1")
			So(err, ShouldBeNil)
			So(u.Scheme, ShouldEqual, "tcp")
			So(u.Host, ShouldEqual, "[2001:db8::1]")
			So(u.Hostname(), ShouldEqual, "2001:db8::1")
			So(u.Port(), ShouldEqual, "")
		})

		Convey("keeps bracketed ipv6 host with port", func() {
			u, err := ParseAddress("[2001:db8::1]:443")
			So(err, ShouldBeNil)
			So(u.Scheme, ShouldEqual, "tcp")
			So(u.Host, ShouldEqual, "[2001:db8::1]:443")
			So(u.Hostname(), ShouldEqual, "2001:db8::1")
			So(u.Port(), ShouldEqual, "443")
		})
	})
}
