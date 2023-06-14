package vertex

const (
	TokenLabel = "Token"
)

type Token struct {
}

func (v Token) Label() string {
	return TokenLabel
}
