package storage

type Storage interface {
	MakeShort(original string) (string, error)
	GetURL(id string) (string, bool)
	ForceSet(id, url string)
}
