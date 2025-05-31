package storage

type Repository interface {
	MakeShort(original string) string
	GetURL(id string) (string, bool)
}
