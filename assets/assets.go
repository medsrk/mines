package assets

import "embed"

//go:embed *.png audio/*.ogg
var FS embed.FS
