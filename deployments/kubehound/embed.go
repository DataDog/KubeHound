package embedconfigdocker

import (
	"embed"
)

const (
	DefaultComposePath = "docker-compose.embed.yaml"
)

//go:embed docker-compose.embed.yaml
var F embed.FS
