package cobra

type Decryptor interface {
	DecryptFlags([]string) ([]string, error)
}
