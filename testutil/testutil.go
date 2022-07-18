package testutil

import (
	"context"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewTestDB(t *testing.T) *gorm.DB {
	file := path.Join(t.TempDir(), "test.db")

	db, err := gorm.Open(sqlite.Open(file), &gorm.Config{})
	assert.Nil(t, err)
	return db
}

type CryptorMock struct {
	mock.Mock
}

func (m *CryptorMock) Encrypt(ctx context.Context, plaintext []byte) ([]byte, error) {
	args := m.Called(ctx, plaintext)
	ciphertext := args.Get(0)
	if ciphertext == nil {
		return nil, args.Error(1)
	}
	return ciphertext.([]byte), args.Error(1)
}

func (m *CryptorMock) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
	args := m.Called(ctx, ciphertext)
	plaintext := args.Get(0)
	if plaintext == nil {
		return nil, args.Error(1)
	}
	return plaintext.([]byte), args.Error(1)
}
