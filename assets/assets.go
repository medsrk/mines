package assets

import "embed"

//go:embed images/*.png audio/*.ogg
var FS embed.FS
