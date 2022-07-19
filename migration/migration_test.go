package migration

import (
	"context"
	"reflect"
	"sort"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/cybercryptio/d1-gorm/testutil"
)

type TestData struct {
	DocumentID     int
	Text           string
	Bytes          []byte
	EncryptedText  string
	EncryptedBytes []byte
}

// The initial schema without encrypted text
type DocumentV1 struct {
	gorm.Model
	DocumentID int
	Text       string
	Bytes      []byte
}

// The new schema with encrypted text
type DocumentV2 struct {
	gorm.Model
	DocumentID     int
	Text           string
	Bytes          []byte
	EncryptedText  string `gorm:"serializer:D1"`
	EncryptedBytes []byte `gorm:"serializer:D1"`
}

// A fake schema used to read the encrypted text without going through the serializer
type DocumentRaw struct {
	gorm.Model
	Text           string
	Bytes          []byte
	EncryptedText  string
	EncryptedBytes []byte
}

func generateTestData(t *testing.T, count int) []TestData {
	testData := make([]TestData, count)
	for i := range testData {
		bytes, err := uuid.New().MarshalBinary()
		assert.Nil(t, err)

		encryptedBytes, err := uuid.New().MarshalBinary()
		assert.Nil(t, err)

		testData[i] = TestData{
			DocumentID:     i,
			Text:           uuid.New().String(),
			EncryptedText:  uuid.New().String(),
			Bytes:          bytes,
			EncryptedBytes: encryptedBytes,
		}
	}
	return testData
}

type SerializerMock struct {
	mock.Mock
}

func (m *SerializerMock) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (interface{}, error) {
	args := m.Called(ctx, field, dst, fieldValue)
	return args.Get(0), args.Error(1)
}

func (m *SerializerMock) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) error {
	args := m.Called(ctx, field, dst, dbValue)
	return args.Error(0)
}

func (m *SerializerMock) OnValue(fieldValue interface{}, dbValue interface{}) *mock.Call {
	return m.On("Value", mock.Anything, mock.Anything, mock.Anything, fieldValue).Return(dbValue, nil)
}

func (m *SerializerMock) OnScan(dbValue interface{}, fieldValue interface{}) *mock.Call {
	return m.On("Scan", mock.Anything, mock.Anything, mock.Anything, dbValue).Return(nil).Run(func(args mock.Arguments) {
		ctx := args.Get(0).(context.Context)
		field := args.Get(1).(*schema.Field)
		dst := args.Get(2).(reflect.Value)
		_ = field.Set(ctx, dst, fieldValue)
	})
}

func TestMigration(t *testing.T) {
	const count = 10
	testData := generateTestData(t, count)

	// Set up the mock serializer
	serializer := &SerializerMock{}
	// We expect two deserializations per entry during the migration.
	// One for the EncryptedText, another for the EncryptedBytes.
	// This is because the migration tries to read the empty data.
	serializer.OnScan(nil, nil).Times(count * 2)
	for _, data := range testData {
		serializer.OnValue(data.Text, data.EncryptedText).Once()
		serializer.OnValue(data.Bytes, data.EncryptedBytes).Once()
	}

	// Set up the database
	schema.RegisterSerializer("D1", serializer)
	db := testutil.NewTestDB(t)

	// Load initial data
	{
		type Document DocumentV1
		err := db.AutoMigrate(&Document{})
		assert.Nil(t, err)

		docs := make([]Document, len(testData))
		for i, d := range testData {
			docs[i] = Document{DocumentID: d.DocumentID, Text: d.Text, Bytes: d.Bytes}
		}

		result := db.Create(&docs)
		assert.Nil(t, result.Error)
	}

	// Do a migration
	{
		type Document DocumentV2
		err := db.AutoMigrate(&Document{})
		assert.Nil(t, err)

		migrateDocument := func(d *Document) {
			d.EncryptedText = d.Text
			d.EncryptedBytes = d.Bytes
		}

		result := Migrate(db, migrateDocument)
		assert.Nil(t, result.Error)
	}

	serializer.AssertExpectations(t)

	// Read the raw data back and verify
	{
		type Document DocumentRaw

		var docs []TestData
		result := db.Model(&Document{}).Find(&docs)
		assert.Nil(t, result.Error)

		assert.Equal(t, len(testData), len(docs))

		sort.Slice(docs, func(i, j int) bool {
			return docs[i].DocumentID < docs[j].DocumentID
		})

		assert.Equal(t, testData, docs)
	}
}
