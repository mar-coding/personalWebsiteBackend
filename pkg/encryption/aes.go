package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"errors"
)

type AES[T any] struct {
	cipher cipher.Block
}

func NewAES[T any](key string) (*AES[T], error) {
	c, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	return &AES[T]{
		cipher: c,
	}, nil
}

func (a *AES[T]) Encrypt(data T) ([]byte, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	out := make([]byte, 16)
	a.cipher.Encrypt(out, b)
	return out, nil
}

func (a *AES[T]) Decrypt(data []byte) (result T, err error) {
	if len(data) == 0 {
		return result, errors.New("input data is empty")
	}
	out := make([]byte, 16)
	a.cipher.Decrypt(out, data)
	if err = json.Unmarshal(out, &result); err != nil {
		return result, err
	}
	return result, nil
}
