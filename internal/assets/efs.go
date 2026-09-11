// Package assets contains the static files like CSS and JavaScript needed to serve the web page
package assets

import (
	"embed"
)

// Templates contains the go html templates
//
//go:embed html
var Templates embed.FS

// Static contains images, CSS, JavaScript
//
//go:embed static
var Static embed.FS
