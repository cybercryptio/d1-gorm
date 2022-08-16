// Copyright 2022 CYBERCRYPT
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License

package testutil

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Utility function that creates an in-memory database to be used for testing.
func NewTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.Nil(t, err)
	return db
}

// CryptorMock is a mock implementation of the Cryptor interface.
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
