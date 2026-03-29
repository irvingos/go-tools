package password

type Hasher interface {
	Algorithm() string
	Hash(password string) (hash, algo string, err error)
	Compare(hash string, password string) (bool, error)
}
