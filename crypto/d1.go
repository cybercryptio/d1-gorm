package crypto

import (
	"context"
	"fmt"

	client "github.com/cybercryptio/d1-client-go/d1-generic"
	pbgeneric "github.com/cybercryptio/d1-client-go/d1-generic/protobuf/generic"
)

// Cryptor is an iterface that abstracts the encryption and decryption of data.
type Cryptor interface {
	Encrypt(ctx context.Context, plaintext []byte) ([]byte, error)
	Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error)
}

// D1Cryptor is an implementation of the Cryptor interface that uses the D1 Generic Service to encrypt and decrypt data.
type D1Cryptor struct {
	d1Client client.GenericClient
}

// NewD1Cryptor creates a new D1Cryptor instance that uses the provided client to connect to the D1 Generic Service. All the database queries across
// all the connections will use this client to encrypt and decrypt data, when necessary.
func NewD1Cryptor(d1Client client.GenericClient) D1Cryptor {
	return D1Cryptor{d1Client: d1Client}
}

// Encrypt calls the D1 Generic Service to encrypt the provided plaintext and returns the concatenation of object ID + ciphertext to be stored in the
// database.
func (c D1Cryptor) Encrypt(ctx context.Context, plaintext []byte) ([]byte, error) {
	res, err := c.d1Client.Generic.Encrypt(ctx, &pbgeneric.EncryptRequest{Plaintext: plaintext})
	if err != nil {
		return nil, err
	}
	return append([]byte(res.ObjectId), res.Ciphertext...), nil
}

// The ciphertext stored in the database is a concatenation of the object ID (of length UUIDLength) and the actual ciphertext.
const UUIDLength = 36

// ErrInvalidFormat is returned when the ciphertext is not in the correct format (object ID of UUIDLength + ciphertext).
var ErrInvalidFormat = fmt.Errorf("the format of the ciphertext is invalid")

// Decrypt parses the database ciphertext to extract the object ID and calls the D1 Generic Service to decrypt the ciphertext and return the
// plaintext.
func (c D1Cryptor) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < UUIDLength {
		return nil, ErrInvalidFormat
	}

	res, err := c.d1Client.Generic.Decrypt(ctx, &pbgeneric.DecryptRequest{
		ObjectId:   string(ciphertext[:UUIDLength]),
		Ciphertext: ciphertext[UUIDLength:],
	})
	if err != nil {
		return nil, err
	}

	return res.Plaintext, nil
}
