package migration

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"sort"
	"testing"

	d1gorm "github.com/cybercryptio/d1-gorm"
	"github.com/cybercryptio/d1-gorm/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// Prepare data
type TestData struct {
	DocumentId     int
	Text           string
	Bytes          []byte
	EncryptedText  string
	EncryptedBytes []byte
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

func generateTestData(count uint) []TestData {
	testData := make([]TestData, count)
	for i := range testData {
		text := fmt.Sprintf("Secret #%d", i)
		encryptedText := fmt.Sprintf("Encrypted [%s]", text)
		bytes := []byte(text)
		encryptedBytes := []byte(encryptedText)
		testData[i] = TestData{i, text, bytes, encryptedText, encryptedBytes}
	}
	return testData
}

func mapSlice[T any, U any](slice []T, mapper func(T) U) []U {
	mapped := make([]U, len(slice))
	for i, value := range slice {
		mapped[i] = mapper(value)
	}
	return mapped
}

func newSorter[T any](sorter func(T, T) bool) func([]T) {
	return func(slice []T) {
		sort.SliceStable(slice, func(i, j int) bool {
			return sorter(slice[i], slice[j])
		})
	}
}

func TestMigration(t *testing.T) {
	const count uint = 10

	testData := generateTestData(count)

	// Set up the mock
	cryptor := &testutil.CryptorMock{}
	for _, data := range testData {
		cryptor.On("Encrypt", mock.Anything, data.Bytes).Once().Return(data.EncryptedBytes, nil)
	}

	// Set up the database
	d1Serializer := d1gorm.NewD1Serializer(cryptor)
	schema.RegisterSerializer("D1", d1Serializer)
	db := testutil.NewTestDB(t)

	// Load initial data
	{
		type Document DocumentV1

		db.AutoMigrate(&Document{})

		docs := mapSlice(testData, func(d TestData) Document {
			return Document{DocumentId: d.DocumentId, Text: d.Text}
		})

		result := db.Create(&docs)
		assert.Nil(t, result.Error)
	}

	// Do a migration
	{
		// The new schema with encrypted text
		type Document DocumentV2

		db.AutoMigrate(&Document{})

		migrateDocument := func(d *Document) {
			d.EncryptedText = d.Text
		}

		result := Migrate(db, migrateDocument)
		assert.Nil(t, result.Error)
	}

	cryptor.AssertExpectations(t)

	// Read the raw data back and verify
	{
		// F

		type Document DocumentRaw

		var docs []Document
		result := db.Find(&docs)
		assert.Nil(t, result.Error)

		type Data struct {
			DocumentId    int
			Text          string
			EncryptedText string
		}

		expected := mapSlice(testData, func(d TestData) Data {
			return Data{
				DocumentId:    d.DocumentId,
				Text:          d.Text,
				EncryptedText: base64.StdEncoding.EncodeToString(d.EncryptedBytes),
			}
		})

		actual := mapSlice(docs, func(d Document) Data {
			return Data{
				DocumentId:    d.DocumentId,
				Text:          d.Text,
				EncryptedText: d.EncryptedText,
			}
		})

		sortData := newSorter(func(a, b Data) bool {
			return a.DocumentId < b.DocumentId
		})

		sortData(expected)
		sortData(actual)

		if !reflect.DeepEqual(expected, actual) {
			t.Fatal("Expected data did not match actual data.")
		}
	}
}
