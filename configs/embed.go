package embedconfig

import (
	"embed"
)

const (
	DefaultPath = "etc/kubehound.yaml"
)

//go:embed etc/*.yaml
var F embed.FS
