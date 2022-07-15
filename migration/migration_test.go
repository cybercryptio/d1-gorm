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
	DocumentId    int
	Text          string
	EncryptedText string
}

// The initial schema without encrypted text
type DocumentV1 struct {
	gorm.Model
	DocumentId int
	Text       string
}

// The new schema with encrypted text
type DocumentV2 struct {
	gorm.Model
	DocumentId    int
	Text          string
	EncryptedText string `gorm:"serializer:D1"`
}

// A fake schema used to read the encrypted text without going through the serializer
type DocumentRaw struct {
	gorm.Model
	DocumentId    int
	Text          string
	EncryptedText string
}

func generateTestData(count int) []TestData {
	testData := make([]TestData, count)
	for i := range testData {
		text := uuid.New().String()
		bytes := ([]byte)(text)
		for i, b := range bytes {
			bytes[i] = b + 1
		}
		encryptedText := string(bytes)
		testData[i] = TestData{
			DocumentId:    i,
			Text:          text,
			EncryptedText: encryptedText,
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
		field.Set(ctx, dst, fieldValue)
	})
}

func TestMigration(t *testing.T) {
	const count = 10
	testData := generateTestData(count)

	// Set up the mock serializer
	serializer := &SerializerMock{}
	serializer.OnScan(nil, nil).Times(count)
	for _, data := range testData {
		serializer.OnValue(data.Text, data.EncryptedText).Once()
	}

	// Set up the database
	schema.RegisterSerializer("D1", serializer)
	db := testutil.NewTestDB(t)

	// Load initial data
	{
		type Document DocumentV1
		db.AutoMigrate(&Document{})

		docs := make([]Document, len(testData))
		for i, d := range testData {
			docs[i] = Document{DocumentId: d.DocumentId, Text: d.Text}
		}

		result := db.Create(&docs)
		assert.Nil(t, result.Error)
	}

	// Do a migration
	{
		type Document DocumentV2
		db.AutoMigrate(&Document{})

		migrateDocument := func(d *Document) {
			d.EncryptedText = d.Text
		}

		result := Migrate(db, migrateDocument)
		assert.Nil(t, result.Error)
	}

	serializer.AssertExpectations(t)

	// Read the raw data back and verify
	{
		type Document DocumentRaw

		var docs []Document
		result := db.Find(&docs)
		assert.Nil(t, result.Error)

		assert.Equal(t, len(testData), len(docs))

		sort.Slice(docs, func(i, j int) bool {
			return docs[i].DocumentId < docs[j].DocumentId
		})

		for i, doc := range docs {
			d := testData[i]
			assert.Equal(t, d.DocumentId, doc.DocumentId)
			assert.Equal(t, d.Text, doc.Text)
			assert.Equal(t, d.EncryptedText, doc.EncryptedText)
		}
	}
}
