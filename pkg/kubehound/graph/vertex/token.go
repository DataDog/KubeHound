package vertex

const (
	tokenLabel = "Pod"
)

type Token struct {
}

func (v Token) Label() string {
	return tokenLabel
}
