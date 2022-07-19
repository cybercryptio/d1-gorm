package crypto

import (
	"context"
	"fmt"

	client "github.com/cybercryptio/d1-client-go/d1-generic"
	pbgeneric "github.com/cybercryptio/d1-client-go/d1-generic/protobuf/generic"
)

type Cryptor interface {
	Encrypt(ctx context.Context, plaintext []byte) ([]byte, error)
	Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error)
}

type D1Cryptor struct {
	d1Client client.GenericClient
}

func NewD1Cryptor(d1Client client.GenericClient) D1Cryptor {
	return D1Cryptor{d1Client: d1Client}
}

func (c D1Cryptor) Encrypt(ctx context.Context, plaintext []byte) ([]byte, error) {
	res, err := c.d1Client.Generic.Encrypt(ctx, &pbgeneric.EncryptRequest{Plaintext: plaintext})
	if err != nil {
		return nil, err
	}
	return append([]byte(res.ObjectId), res.Ciphertext...), nil
}

const UUID_LENGTH = 36

var ErrInvalidFormat = fmt.Errorf("the format of the ciphertext is invalid")

func (c D1Cryptor) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < UUID_LENGTH {
		return nil, ErrInvalidFormat
	}

	res, err := c.d1Client.Generic.Decrypt(ctx, &pbgeneric.DecryptRequest{
		ObjectId:   string(ciphertext[:UUID_LENGTH]),
		Ciphertext: ciphertext[UUID_LENGTH:],
	})
	if err != nil {
		return nil, err
	}

	return res.Plaintext, nil
}
