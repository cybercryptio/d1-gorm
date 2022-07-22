package d1gorm

import (
	"context"
	"encoding/base64"
	"fmt"
	"reflect"

	"github.com/cybercryptio/d1-gorm/crypto"
	"gorm.io/gorm/schema"
)

// Error returned when trying to encrypt a field of an unsupported type.
var ErrEncryptUnsupported = fmt.Errorf("supported encryption field types: string, []byte")

// Error returned when trying to decrypt a field of an unsupported type.
var ErrDecryptUnsupported = fmt.Errorf("supported decryption field types: string, []byte")

// D1Serializer is used to transparently encrypt and decrypt data when reading/writing to the database. To use it you must instantiate it, register it
// to be used for your gorm schema with schema.RegisterSerializer("D1", d1Serializer), and tag the model fields to be serialized with
// `gorm:"serializer:D1"`. Currently only the serialization of string and []byte data types is supported.
type D1Serializer struct {
	cryptor crypto.Cryptor
}

// NewD1Serializer creates a new D1Serializer that uses the provided Cryptor to encrypt and decrypt data.
func NewD1Serializer(cryptor crypto.Cryptor) D1Serializer {
	return D1Serializer{cryptor: cryptor}
}

// Value is called by gorm to serialize the value of a field before being written to the database.
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

// Scan is called by gorm to deserialize the value of a field after it has been read from the database.
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
		return field.Set(ctx, dst, nil)
	default:
		return fmt.Errorf("decryption of type %T: %w", value, ErrDecryptUnsupported)
	}

	decryptedValue, err := s.cryptor.Decrypt(ctx, valueBytes)
	if err != nil {
		return err
	}

	return field.Set(ctx, dst, decryptedValue)
}
