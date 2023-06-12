package vertex

const (
	tokenLabel = "Token"
)

type Token struct {
}

func (v Token) Label() string {
	return tokenLabel
}
