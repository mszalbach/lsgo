package assets

import (
	"embed"
)

//go:embed html
var Templates embed.FS

//go:embed static
var Static embed.FS
