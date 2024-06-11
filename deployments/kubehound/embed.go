package embedconfigdocker

import (
	"embed"
)

//go:embed *.yaml
//go:embed *.yaml.tpl
var F embed.FS
