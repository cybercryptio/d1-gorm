package d1gorm

import (
	"context"
	"encoding/base64"
	"fmt"
	"reflect"

	"github.com/cybercryptio/d1-gorm/encrypt"
	"gorm.io/gorm/schema"
)

var ErrEncryptUnsupported = fmt.Errorf("supported encryption field types: string, []byte")
var ErrDecryptUnsupported = fmt.Errorf("supported decryption field types: string, []byte")

type D1Serializer struct {
	cryptor encrypt.Cryptor
}

func NewD1Serializer(cryptor encrypt.Cryptor) D1Serializer {
	return D1Serializer{cryptor: cryptor}
}

func (s D1Serializer) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (interface{}, error) {
	switch value := fieldValue.(type) {
	case []byte:
		encryptedValue, err := s.cryptor.Encrypt(ctx, value)
		if err != nil {
			return nil, err
		}
		return encryptedValue, nil
	case string:
		encryptedValue, err := s.cryptor.Encrypt(ctx, []byte(value))
		if err != nil {
			return nil, err
		}
		return base64.StdEncoding.EncodeToString(encryptedValue), nil
	case nil:
		return nil, nil
	default:
		return nil, fmt.Errorf("encryption of type %T: %w", value, ErrEncryptUnsupported)
	}
}

func (s D1Serializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) error {
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
		return fmt.Errorf("decryption of type %T: %w", value, ErrDecryptUnsupported)
	}

	decryptedValue, err := s.cryptor.Decrypt(ctx, valueBytes)
	if err != nil {
		return err
	}

	field.Set(ctx, dst, decryptedValue)

	return nil
}
