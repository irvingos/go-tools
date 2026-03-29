package password

import "golang.org/x/crypto/bcrypt"

type bcryptHasher struct {
	cost int
}

// Algorithm implements [Hasher].
func (b *bcryptHasher) Algorithm() string {
	return "bcrypt"
}

// Hash implements [Hasher].
func (b *bcryptHasher) Hash(password string) (hash, algo string, err error) {
	encrypt, err := bcrypt.GenerateFromPassword([]byte(password), b.cost)
	if err != nil {
		return
	}
	return string(encrypt), b.Algorithm(), nil
}

// Compare implements [Hasher].
func (b *bcryptHasher) Compare(hash string, password string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err == nil {
		return true, nil
	}
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return false, nil
	}
	return false, err
}

func NewBcryptHasher(cost int) Hasher {
	return &bcryptHasher{
		cost: cost,
	}
}
