package static

import "embed"

//go:embed *.html css images js vendor
var Files embed.FS
