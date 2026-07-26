package ports

type Store interface {
	Exists(path string) bool
	Write(path string, content string) error
	Append(path string, content string) error
}
