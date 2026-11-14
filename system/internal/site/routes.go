package site

import "strings"

// pageRoute maps HTML filenames to their canonical public URL.
func pageRoute(file string) string {
	route := "/" + strings.TrimSuffix(file, ".html")
	if route == "/index" {
		return "/"
	}
	return strings.TrimSuffix(route, "/index")
}
