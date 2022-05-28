package ping

import (
	pkgurl "net/url"
	"strings"
)

// ParseAddress will try to parse addr as url.URL.
func ParseAddress(addr string) (*pkgurl.URL, error) {
	if strings.Contains(addr, "://") {
		// it maybe with scheme, try url.Parse
		return pkgurl.Parse(addr)
	}
	return pkgurl.Parse("tcp://" + addr)
}
