package convert

import "strings"

func CanonicalPageRoutePath(p string) string {
	if p == "" {
		return p
	}
	if strings.HasPrefix(p, "/admin/") {
		return strings.TrimPrefix(p, "/admin")
	}
	if p == "/admin" {
		return "/"
	}
	return p
}

