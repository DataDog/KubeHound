package config

const (
	DefaultSQLitePath = "kubehound.sqlite"

	SQLitePath = "sqlite.path"
)

// SQLiteConfig configures SQLite-specific parameters.
type SQLiteConfig struct {
	Path string `mapstructure:"path"`
}
