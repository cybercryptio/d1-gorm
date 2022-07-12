package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"reflect"

	d1client "github.com/cybercryptio/d1-client-go/d1-generic"
	pbgeneric "github.com/cybercryptio/d1-client-go/d1-generic/protobuf/generic"
	"google.golang.org/grpc/metadata"
	"gorm.io/gorm/schema"
)

type D1Cryptor struct {
	d1Client     d1client.GenericClient
	tokenFactory func(context.Context) (string, error)
}

func NewD1Cryptor(d1Client d1client.GenericClient, tokenFactory func(context.Context) (string, error)) D1Cryptor {
	return D1Cryptor{d1Client: d1Client, tokenFactory: tokenFactory}
}

func (c D1Cryptor) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (interface{}, error) {
	switch value := fieldValue.(type) {
	case []byte:
		encryptedValue, err := encryptBytes(ctx, c, value)
		if err != nil {
			return nil, err
		}
		return encryptedValue, nil
	case string:
		encryptedValue, err := encryptBytes(ctx, c, []byte(value))
		if err != nil {
			return nil, err
		}
		return base64.StdEncoding.EncodeToString(encryptedValue), nil
	case nil:
		return nil, nil
	default:
		return nil, fmt.Errorf("encryption of type %#v is not supported; only string and []byte are supported", value)
	}
}

func (c D1Cryptor) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) error {
	var valueBytes []byte
	var err error

	switch value := dbValue.(type) {
	case []byte:
		valueBytes = value
	case string:
		valueBytes, err = base64.StdEncoding.DecodeString(value)
		if err != nil {
			return err
		}
	case nil:
		field.Set(ctx, dst, nil)
		return nil
	default:
		return fmt.Errorf("decryption of type %#v is not supported; only string and []byte are supported", value)
	}

	decryptedValue, err := decryptBytes(ctx, c, valueBytes)
	if err != nil {
		return err
	}

	field.Set(ctx, dst, decryptedValue)

	return nil
}

const UUID_LENGTH = 36

func encryptBytes(ctx context.Context, c D1Cryptor, plaintext []byte) ([]byte, error) {
	accessToken, err := c.tokenFactory(ctx)
	if err != nil {
		return nil, err
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "bearer "+accessToken)
	res, err := c.d1Client.Generic.Encrypt(ctx, &pbgeneric.EncryptRequest{Plaintext: plaintext})
	if err != nil {
		return nil, err
	}
	return append([]byte(res.ObjectId), res.Ciphertext...), nil
}

func decryptBytes(ctx context.Context, c D1Cryptor, ciphertext []byte) ([]byte, error) {
	accessToken, err := c.tokenFactory(ctx)
	if err != nil {
		return nil, err
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "bearer "+accessToken)
	res, err := c.d1Client.Generic.Decrypt(ctx, &pbgeneric.DecryptRequest{
		ObjectId:   string(ciphertext[:UUID_LENGTH]),
		Ciphertext: ciphertext[UUID_LENGTH:],
	})
	if err != nil {
		return nil, err
	}
	return res.Plaintext, nil
}
