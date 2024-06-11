package embedconfigdocker

import (
	"embed"
)

//go:embed *.yaml
var F embed.FS
