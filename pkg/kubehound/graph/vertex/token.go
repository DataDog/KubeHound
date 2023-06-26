package vertex

const (
	TokenLabel = "Token"
)

// Token is a pseudo-vertex that does not require a builder.
type Token struct {
}

func (v Token) Label() string {
	return TokenLabel
}
