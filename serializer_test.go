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

package d1gorm

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm/schema"

	"github.com/cybercryptio/d1-gorm/testutil"
)

func TestSerializerString(t *testing.T) {
	type PersonString struct {
		FirstName string
		LastName  string `gorm:"serializer:D1"`
	}

	firstName := "John"
	lastName := "Doe"
	encryptedLastName := "Doencrypt"

	cryptor := &testutil.CryptorMock{}
	cryptor.On("Encrypt", mock.Anything, []byte(lastName)).Once().Return([]byte(encryptedLastName), nil)
	cryptor.On("Decrypt", mock.Anything, []byte(encryptedLastName)).Once().Return([]byte(lastName), nil)

	d1Serializer := NewD1Serializer(cryptor)
	schema.RegisterSerializer("D1", d1Serializer)

	db := testutil.NewTestDB(t)
	err := db.AutoMigrate(&PersonString{})
	assert.Nil(t, err)

	err = db.Create(&PersonString{FirstName: firstName, LastName: lastName}).Error
	assert.Nil(t, err)

	p := &PersonString{}
	err = db.Where("first_name = ?", firstName).First(p).Error
	assert.Nil(t, err)

	assert.Equal(t, firstName, p.FirstName)
	assert.Equal(t, lastName, p.LastName)
	cryptor.AssertExpectations(t)
}

func TestSerializerBytes(t *testing.T) {
	type PersonBytes struct {
		FirstName string
		LastName  []byte `gorm:"serializer:D1"`
	}

	firstName := "John"
	lastName := []byte("Doe")
	encryptedLastName := []byte("Doencrypt")

	cryptor := &testutil.CryptorMock{}
	cryptor.On("Encrypt", mock.Anything, lastName).Once().Return(encryptedLastName, nil)
	cryptor.On("Decrypt", mock.Anything, encryptedLastName).Once().Return(lastName, nil)

	d1Serializer := NewD1Serializer(cryptor)
	schema.RegisterSerializer("D1", d1Serializer)

	db := testutil.NewTestDB(t)
	err := db.AutoMigrate(&PersonBytes{})
	assert.Nil(t, err)

	err = db.Create(&PersonBytes{FirstName: firstName, LastName: lastName}).Error
	assert.Nil(t, err)

	p := &PersonBytes{}
	err = db.Where("first_name = ?", firstName).First(p).Error
	assert.Nil(t, err)

	assert.Equal(t, firstName, p.FirstName)
	assert.Equal(t, lastName, p.LastName)
	cryptor.AssertExpectations(t)
}

func TestSerializerNil(t *testing.T) {
	type PersonBytes struct {
		FirstName string
		LastName  []byte `gorm:"serializer:D1"`
	}

	firstName := "John"

	cryptor := &testutil.CryptorMock{}

	d1Serializer := NewD1Serializer(cryptor)
	schema.RegisterSerializer("D1", d1Serializer)

	db := testutil.NewTestDB(t)
	err := db.AutoMigrate(&PersonBytes{})
	assert.Nil(t, err)

	err = db.Create(&PersonBytes{FirstName: firstName}).Error
	assert.Nil(t, err)

	p := &PersonBytes{}
	err = db.Where("first_name = ?", firstName).First(p).Error
	assert.Nil(t, err)

	assert.Equal(t, firstName, p.FirstName)
	assert.Nil(t, p.LastName)
	cryptor.AssertExpectations(t)
}

func TestSerializerUnsupported(t *testing.T) {
	firstName1 := "John"
	firstName2 := "Henry"

	db := testutil.NewTestDB(t)

	{
		type PersonAge struct {
			FirstName string
			Age       int
		}

		err := db.AutoMigrate(&PersonAge{})
		assert.Nil(t, err)

		err = db.Create(&PersonAge{FirstName: firstName1, Age: 30}).Error
		assert.Nil(t, err)
	}

	{
		type PersonAge struct {
			FirstName string
			Age       int `gorm:"serializer:D1"`
		}

		cryptor := &testutil.CryptorMock{}

		d1Serializer := NewD1Serializer(cryptor)
		schema.RegisterSerializer("D1", d1Serializer)

		err := db.AutoMigrate(&PersonAge{})
		assert.Nil(t, err)

		err = db.Create(&PersonAge{FirstName: firstName2, Age: 30}).Error
		assert.ErrorContains(t, err, ErrEncryptUnsupported.Error())

		p := &PersonAge{}
		err = db.Where("first_name = ?", firstName1).First(p).Error
		assert.ErrorContains(t, err, ErrDecryptUnsupported.Error())

		cryptor.AssertExpectations(t)
	}
}

func TestSerializerEncryptError(t *testing.T) {
	type PersonString struct {
		FirstName string
		LastName  string `gorm:"serializer:D1"`
	}

	firstName := "John"
	lastName := "Doe"
	ErrEncrypt := fmt.Errorf("encryption error")

	cryptor := &testutil.CryptorMock{}
	// The Serializer's Value method will be called twice: once when trying to insert into the database, and once for logging the statement that
	// returned the error.
	cryptor.On("Encrypt", mock.Anything, []byte(lastName)).Return(nil, ErrEncrypt)

	d1Serializer := NewD1Serializer(cryptor)
	schema.RegisterSerializer("D1", d1Serializer)

	db := testutil.NewTestDB(t)
	err := db.AutoMigrate(&PersonString{})
	assert.Nil(t, err)

	err = db.Create(&PersonString{FirstName: firstName, LastName: lastName}).Error
	assert.ErrorContains(t, err, ErrEncrypt.Error())
	cryptor.AssertExpectations(t)
}

func TestSerializerDecryptError(t *testing.T) {
	type PersonString struct {
		FirstName string
		LastName  string `gorm:"serializer:D1"`
	}

	firstName := "John"
	lastName := "Doe"
	encryptedLastName := "Doencrypt"
	ErrDecrypt := fmt.Errorf("decryption error")

	cryptor := &testutil.CryptorMock{}
	cryptor.On("Encrypt", mock.Anything, []byte(lastName)).Once().Return([]byte(encryptedLastName), nil)
	cryptor.On("Decrypt", mock.Anything, []byte(encryptedLastName)).Once().Return(nil, ErrDecrypt)

	d1Serializer := NewD1Serializer(cryptor)
	schema.RegisterSerializer("D1", d1Serializer)

	db := testutil.NewTestDB(t)
	err := db.AutoMigrate(&PersonString{})
	assert.Nil(t, err)

	err = db.Create(&PersonString{FirstName: firstName, LastName: lastName}).Error
	assert.Nil(t, err)

	p := &PersonString{}
	err = db.Where("first_name = ?", firstName).First(p).Error
	assert.ErrorContains(t, err, ErrDecrypt.Error())
	cryptor.AssertExpectations(t)
}

type ctxKeyType string

type cryptorTestCtx struct {
	t           *testing.T
	token       string
	tokenCtxKey ctxKeyType
}

func (c *cryptorTestCtx) Encrypt(ctx context.Context, plaintext []byte) ([]byte, error) {
	token, found := ctx.Value(c.tokenCtxKey).(string)
	assert.True(c.t, found)
	assert.Equal(c.t, c.token, token)
	return plaintext, nil
}

func (c *cryptorTestCtx) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
	token, found := ctx.Value(c.tokenCtxKey).(string)
	assert.True(c.t, found)
	assert.Equal(c.t, c.token, token)
	return ciphertext, nil
}

func TestSerializerPassTokenInCtx(t *testing.T) {
	type PersonString struct {
		FirstName string
		LastName  string `gorm:"serializer:D1"`
	}

	firstName := "John"
	lastName := "Doe"
	token := "token"
	tokenCtxKey := ctxKeyType("tokenCtxKey")

	cryptor := &cryptorTestCtx{t, token, tokenCtxKey}

	d1Serializer := NewD1Serializer(cryptor)
	schema.RegisterSerializer("D1", d1Serializer)

	db := testutil.NewTestDB(t)
	err := db.AutoMigrate(&PersonString{})
	assert.Nil(t, err)

	ctx := context.WithValue(context.Background(), tokenCtxKey, token)

	err = db.WithContext(ctx).Create(&PersonString{FirstName: firstName, LastName: lastName}).Error
	assert.Nil(t, err)

	p := &PersonString{}
	err = db.WithContext(ctx).Where("first_name = ?", firstName).First(p).Error
	assert.Nil(t, err)
	assert.Equal(t, firstName, p.FirstName)
	assert.Equal(t, lastName, p.LastName)
}
