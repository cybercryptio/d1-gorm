package d1gorm

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type CryptorMock struct {
	mock.Mock
}

func (m *CryptorMock) Encrypt(ctx context.Context, plaintext []byte) ([]byte, error) {
	args := m.Called(ctx, plaintext)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *CryptorMock) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
	args := m.Called(ctx, ciphertext)
	return args.Get(0).([]byte), args.Error(1)
}

func TestSerializerString(t *testing.T) {
	type PersonString struct {
		FirstName string
		LastName  string `gorm:"serializer:D1"`
	}

	firstName := "John"
	lastName := "Doe"
	encryptedLastName := "Doencrypt"

	cryptor := &CryptorMock{}
	cryptor.On("Encrypt", mock.Anything, []byte(lastName)).Once().Return([]byte(encryptedLastName), nil)
	cryptor.On("Decrypt", mock.Anything, []byte(encryptedLastName)).Once().Return([]byte(lastName), nil)

	d1Serializer := NewD1Serializer(cryptor)
	schema.RegisterSerializer("D1", d1Serializer)

	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	assert.Nil(t, err)
	db.AutoMigrate(&PersonString{})

	err = db.Create(&PersonString{FirstName: firstName, LastName: lastName}).Error
	assert.Nil(t, err)

	p := &PersonString{}
	err = db.Where("first_name = ?", firstName).First(p).Error
	assert.Nil(t, err)

	assert.Equal(t, firstName, p.FirstName)
	assert.Equal(t, lastName, p.LastName)
	cryptor.AssertExpectations(t)

	os.Remove("test.db")
}

func TestSerializerBytes(t *testing.T) {
	type PersonBytes struct {
		FirstName string
		LastName  []byte `gorm:"serializer:D1"`
	}

	firstName := "John"
	lastName := []byte("Doe")
	encryptedLastName := []byte("Doencrypt")

	cryptor := &CryptorMock{}
	cryptor.On("Encrypt", mock.Anything, lastName).Once().Return(encryptedLastName, nil)
	cryptor.On("Decrypt", mock.Anything, encryptedLastName).Once().Return(lastName, nil)

	d1Serializer := NewD1Serializer(cryptor)
	schema.RegisterSerializer("D1", d1Serializer)

	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	assert.Nil(t, err)
	db.AutoMigrate(&PersonBytes{})

	err = db.Create(&PersonBytes{FirstName: firstName, LastName: lastName}).Error
	assert.Nil(t, err)

	p := &PersonBytes{}
	err = db.Where("first_name = ?", firstName).First(p).Error
	assert.Nil(t, err)

	assert.Equal(t, firstName, p.FirstName)
	assert.Equal(t, lastName, p.LastName)
	cryptor.AssertExpectations(t)

	os.Remove("test.db")
}

func TestSerializerNil(t *testing.T) {
	type PersonBytes struct {
		FirstName string
		LastName  []byte `gorm:"serializer:D1"`
	}

	firstName := "John"

	cryptor := &CryptorMock{}

	d1Serializer := NewD1Serializer(cryptor)
	schema.RegisterSerializer("D1", d1Serializer)

	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	assert.Nil(t, err)
	db.AutoMigrate(&PersonBytes{})

	err = db.Create(&PersonBytes{FirstName: firstName}).Error
	assert.Nil(t, err)

	p := &PersonBytes{}
	err = db.Where("first_name = ?", firstName).First(p).Error
	assert.Nil(t, err)

	assert.Equal(t, firstName, p.FirstName)
	assert.Nil(t, p.LastName)
	cryptor.AssertExpectations(t)

	os.Remove("test.db")
}

func TestSerializerUnsupported(t *testing.T) {
	firstName1 := "John"
	firstName2 := "Henry"

	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	assert.Nil(t, err)

	{
		type PersonAge struct {
			FirstName string
			Age       int
		}

		db.AutoMigrate(&PersonAge{})

		err = db.Create(&PersonAge{FirstName: firstName1, Age: 30}).Error
		assert.Nil(t, err)
	}

	{
		type PersonAge struct {
			FirstName string
			Age       int `gorm:"serializer:D1"`
		}

		cryptor := &CryptorMock{}

		d1Serializer := NewD1Serializer(cryptor)
		schema.RegisterSerializer("D1", d1Serializer)

		db.AutoMigrate(&PersonAge{})

		err = db.Create(&PersonAge{FirstName: firstName2, Age: 30}).Error
		assert.ErrorContains(t, err, ErrEncryptUnsupported.Error())

		p := &PersonAge{}
		err = db.Where("first_name = ?", firstName1).First(p).Error
		assert.ErrorContains(t, err, ErrDecryptUnsupported.Error())

		cryptor.AssertExpectations(t)
	}

	os.Remove("test.db")
}
