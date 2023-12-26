package config

const (
	DefaultIngestorAPIPort = 9000
	DefaultIngestorAPIAddr = "127.0.0.1"
)

type IngestorConfig struct {
	Addr string
	Port int
}
