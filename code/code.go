package code

import (
	"crypto/rand"
	"math/big"
)

var base62Chars = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")

func Generate(length int) (string, error) {
	token := make([]rune, length)
	for i := range token {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(base62Chars))))
		if err != nil {
			return "", err
		}
		token[i] = base62Chars[n.Int64()]
	}
	return string(token), nil
}
