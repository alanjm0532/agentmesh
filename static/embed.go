package static

import "embed"

//go:embed index.html css images js vendor
var Files embed.FS
