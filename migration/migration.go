package migration

import (
	"fmt"
	"math/rand"
	"time"

	"gorm.io/gorm"
)

func initUsers(db *gorm.DB, count int) {
	type User struct {
		gorm.Model
		FirstName string
		LastName  string
	}

	db.Migrator().DropTable(&User{})

	db.AutoMigrate(&User{})

	users := make([]User, count)

	seed := time.Now().UTC().UnixNano()
	rand.Seed(seed)

	for i := range users {
		users[i] = User{
			FirstName: fmt.Sprintf("John%d", rand.Intn(1000)),
			LastName:  fmt.Sprintf("Doe%d", rand.Intn(1000)),
		}
	}

	db.Create(&users)
}

func Migrate[T any](db *gorm.DB, migrate func(*T), opts ...option) {
	o := options{
		batchSize: 10,
	}
	for _, opt := range opts {
		opt(&o)
	}

	var value T
	reader := db.Model(&value)
	if o.debug {
		reader = reader.Debug()
	}
	if len(o.readFields) > 0 {
		reader = reader.Select(o.readFields)
	}

	entries := make([]T, o.batchSize)
	reader.FindInBatches(&entries, o.batchSize, func(tx *gorm.DB, batch int) error {
		if o.debug {
			fmt.Printf("Processing batch #%d (%d elements)", batch, len(entries))
		}

		for i := range entries {
			migrate(&entries[i])
		}

		writer := tx
		if o.debug {
			writer = writer.Debug()
		}
		if len(o.writeFields) > 0 {
			writer = writer.Omit("created_at").Select(o.writeFields)
		}
		writer.Save(&entries)

		return nil
	})
}

func TryMigrate(db *gorm.DB) {
	initUsers(db, 10)

	type User struct {
		gorm.Model

		// The old, unencrypted first and last name
		FirstName string
		LastName  string

		// The new, encrypted first and last name
		FirstNameEncrypted string `gorm:"serializer:D1"`
		LastNameEncrypted  string `gorm:"serializer:D1"`
	}

	migrateUser := func(u *User) {
		u.FirstNameEncrypted = u.FirstName
		u.LastNameEncrypted = u.LastName
	}

	db.AutoMigrate(&User{})
	Migrate(db, migrateUser,
		Read("id", "first_name", "last_name"),
		Write("id", "first_name_encrypted", "last_name_encrypted"),
		//BatchSize(4),
		Debug(true),
	)
}
