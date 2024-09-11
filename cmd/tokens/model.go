package tokens

type Token struct {
	ID              uint
	AccountUsername string
	Token           string
	Type            string
	State           string
	CreatedAt       uint64
}
