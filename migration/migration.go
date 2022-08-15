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

package migration

import (
	"gorm.io/gorm"
)

// Migrate can be used for migrating data from an existing database with unencrypted data to a new format which has encrypted data in the colums
// tagged as such. Before calling migrate, the user has to extend the database tables with columns for holding the encrypted data. Then Migrate can be
// called on the database with a function that works on the migrated data type and which will be executed to migrate the database entries. For more
// details, see the example in the godoc.
func Migrate[T any](db *gorm.DB, migrate func(*T), opts ...Option) *gorm.DB {
	o := defaultOptions()
	o.apply(opts...)

	session := db.Session(&gorm.Session{QueryFields: true})
	var entries []T
	result := session.FindInBatches(&entries, o.batchSize, func(tx *gorm.DB, batch int) error {
		for i := range entries {
			migrate(&entries[i])
		}

		tx.Save(&entries)

		return nil
	})
	return result
}
