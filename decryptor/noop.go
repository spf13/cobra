package decryptor

type NoopDecryptor struct{}

func NewNoopDecryptor() Decryptor {
	return &NoopDecryptor{}
}

func (n *NoopDecryptor) DecryptArguments(args []string) ([]string, error) {
	return args, nil
}
