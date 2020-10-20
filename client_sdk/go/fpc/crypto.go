package fpc

func Encrypt(input []byte, encryptionKey []byte) ([]byte, error) {
	return input, nil
}

func KeyGen() ([]byte, error) {
	return []byte("fake key"), nil
}

func Decrypt(encryptedResponse []byte, resultEncryptionKey []byte) ([]byte, error) {
	return encryptedResponse, nil
}
