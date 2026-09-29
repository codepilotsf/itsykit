package site

import "strings"

// pageRoute is shared by routing and snip ownership. Keep one mapping so a
// directory index cannot acquire a different content scope from its public URL.
func pageRoute(file string) string {
	route := "/" + strings.TrimSuffix(file, ".html")
	if route == "/index" {
		return "/"
	}
	return strings.TrimSuffix(route, "/index")
}
