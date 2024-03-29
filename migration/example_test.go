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

//nolint:gosec // for using math/rand
package migration

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func AddUsers(db *gorm.DB, count int) {
	type User struct {
		gorm.Model
		FirstName     string
		LastName      string
		VideosWatched int
		Favorites     int
	}

	users := make([]User, count)

	firstNames := []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Miller", "Davis", "Garcia", "Rodriguez", "Wilson"}
	lastNames := []string{"James", "Robert", "John", "Michael", "David", "William", "Richard", "Joseph", "Thomas", "Charles"}

	seed := time.Now().UTC().UnixNano()
	rand.Seed(seed)

	for i := range users {
		watched := rand.Intn(1000)
		favorites := rand.Intn(watched)
		users[i] = User{
			FirstName:     firstNames[rand.Intn(len(firstNames))],
			LastName:      lastNames[rand.Intn(len(lastNames))],
			VideosWatched: watched,
			Favorites:     favorites,
		}
	}

	db.Create(&users)
}

func ExampleMigrate() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	if err != nil {
		log.Fatal(err)
	}

	{
		// The initial setup.
		// We define an initial model with encrypted data and create dummy data.
		type User struct {
			gorm.Model
			FirstName     string
			LastName      string
			VideosWatched int
			Favorites     int
		}

		_ = db.Migrator().DropTable(&User{})
		_ = db.AutoMigrate(&User{})

		AddUsers(db, 24)
	}

	{
		// The migration.
		// The model is extended with new colums for encrypted first and last names.
		type User struct {
			gorm.Model
			FirstName          string
			LastName           string
			FirstNameEncrypted string `gorm:"serializer:D1"`
			LastNameEncrypted  string `gorm:"serializer:D1"`
			VideosWatched      int
			Favorites          int
		}

		// This will adjust the database schema to have the two new columns.
		_ = db.AutoMigrate(&User{})

		// The old unencrypted first and last names, should be migrated to the encrypted columns.
		migrateUser := func(u *User) {
			u.FirstNameEncrypted = u.FirstName
			u.LastNameEncrypted = u.LastName
			// You can clear the old data at the same time if you want.
			//u.FirstName = ""
			//u.LastName = ""
		}

		// This scope is used to make sure we only migrate data that is not yet encrypted.
		unencrypted := func(db *gorm.DB) *gorm.DB {
			return db.Where(map[string]interface{}{
				"first_name_encrypted": nil,
				"last_name_encrypted":  nil},
			)
		}

		result := Migrate(db.Scopes(unencrypted), migrateUser)
		fmt.Printf("Rows affected: %d, Error: %v", result.RowsAffected, result.Error)
		// You can also specify an option to set the batch size.
		//Migrate(db, migrateUser, BatchSize(20))

		// Add some more users with unencrypted data and run again.
		AddUsers(db, 18)

		result = Migrate(db.Scopes(unencrypted), migrateUser)
		fmt.Printf("Rows affected: %d, Error: %v", result.RowsAffected, result.Error)
	}
}
