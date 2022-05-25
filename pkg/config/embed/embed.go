package embed

import (
	"embed"
)

//go:embed *
var Conf embed.FS
